package compiler

import (
	"ast"
	"code"
	"fmt"
	"object"
)

type CompilationScope struct {
	instructions     code.Instructions
	localSymbolTable *SymbolTable

	lastOpCodeStartPos       int
	secondLastOpCodeStartPos int
}

type Compiler struct {
	constants []object.Object

	scopes     []CompilationScope
	scopeIndex int
}

func New() *Compiler {
	symbol := NewSymbolTable()

	for i, v := range object.Builtins {
		symbol.DefineBuiltin(i, v.Name)
	}

	mainScope := CompilationScope{
		instructions:             []byte{},
		localSymbolTable:         symbol,
		lastOpCodeStartPos:       0,
		secondLastOpCodeStartPos: 0,
	}
	return &Compiler{scopes: []CompilationScope{mainScope}, scopeIndex: 0, constants: []object.Object{}}
}

func NewWithStates(constants []object.Object, symbolTable *SymbolTable) *Compiler {
	mainScope := CompilationScope{
		instructions:             []byte{},
		localSymbolTable:         symbolTable,
		lastOpCodeStartPos:       0,
		secondLastOpCodeStartPos: 0,
	}

	return &Compiler{scopes: []CompilationScope{mainScope}, scopeIndex: 0, constants: constants}
}

func (c *Compiler) currentScope() *CompilationScope {
	return &c.scopes[c.scopeIndex]
}

func (c *Compiler) currentInstructions() code.Instructions {
	return c.currentScope().instructions
}

func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions:             []byte{},
		localSymbolTable:         NewEnclosedSymbolTable(c.currentScope().localSymbolTable),
		lastOpCodeStartPos:       0,
		secondLastOpCodeStartPos: 0,
	}

	c.scopes = append(c.scopes, scope)
	c.scopeIndex++
}

func (c *Compiler) leaveScope() CompilationScope {
	lastScopeIndex := len(c.scopes) - 1
	lastScope := c.scopes[lastScopeIndex]
	c.scopes = c.scopes[:lastScopeIndex]
	c.scopeIndex--
	return lastScope
}

func (c *Compiler) addConstant(value object.Object) int {
	c.constants = append(c.constants, value)
	return len(c.constants) - 1
}

func (c *Compiler) shiftLastOpCodeStartPos(lastOpCodeStartPos int) {
	secondLast := c.currentScope().lastOpCodeStartPos
	c.currentScope().lastOpCodeStartPos = lastOpCodeStartPos
	c.currentScope().secondLastOpCodeStartPos = secondLast
}

func (c *Compiler) emit(op code.OpCode, operands ...int) int {
	ins := code.Make(op, operands...)
	startPos := len(c.currentInstructions())
	newInstructions := append(c.currentInstructions(), ins...)
	c.currentScope().instructions = newInstructions
	c.shiftLastOpCodeStartPos(startPos)
	return startPos
}

func (c *Compiler) compileInfixExpression(node *ast.InfixExpression) error {
	op := node.Operator
	var err error
	if op == "<" || op == "<=" {
		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		err := c.Compile(node.Left)
		if err != nil {
			return err
		}
	} else {
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}
	}

	switch node.Operator {
	case "[":
		c.emit(code.OpIndex)
	case "+":
		c.emit(code.OpAdd)
	case "-":
		c.emit(code.OpSubtraction)
	case "*":
		c.emit(code.OpMultiply)
	case "/":
		c.emit(code.OpDivide)
	case "==":
		c.emit(code.OpEqual)
	case "!=":
		c.emit(code.OpNotEqual)
	case ">", "<":
		c.emit(code.OpGreaterThan)
	case ">=", "<=":
		c.emit(code.OpGreaterEqual)

	default:
		return fmt.Errorf("unknown operator %s", op)
	}
	return nil
}

func (c *Compiler) replaceOperands(opCodeStartPos int, newOperands ...int) {
	codeToReplace := c.currentInstructions()[opCodeStartPos]
	c.replaceInstructions(opCodeStartPos, code.OpCode(codeToReplace), newOperands...)
}

func (c *Compiler) replaceInstructions(opCodeStartPos int, op code.OpCode, newOperands ...int) {
	newIns := code.Make(code.OpCode(op), newOperands...)
	for i, ins := range newIns {
		c.currentScope().instructions[i+opCodeStartPos] = ins
	}
}

func (c *Compiler) lastOpIs(testOp code.OpCode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}

	lastOpCode := code.OpCode(c.currentInstructions()[c.currentScope().lastOpCodeStartPos])
	return lastOpCode == testOp
}

func (c *Compiler) removeLastOp() {
	c.currentScope().instructions = c.currentInstructions()[:c.currentScope().lastOpCodeStartPos]
	c.currentScope().lastOpCodeStartPos = c.currentScope().secondLastOpCodeStartPos
	c.currentScope().secondLastOpCodeStartPos = -1
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, statement := range node.Statements {
			err := c.Compile(statement)
			if err != nil {
				return err
			}
		}
	case *ast.BlockExpression:
		for _, statement := range node.Statements {
			err := c.Compile(statement)
			if err != nil {
				return err
			}
		}
	case *ast.ExpressionStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)
	case *ast.PrefixExpression:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "-":
			c.emit(code.OpMinus)
		case "!":
			c.emit(code.OpBang)
		default:
			return fmt.Errorf("unsupported prefix operator %s", node.Operator)
		}
	case *ast.InfixExpression:
		err := c.compileInfixExpression(node)
		if err != nil {
			return err
		}
	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		jumpNotTruethyPos := c.emit(code.OpJumptNotTruethy, 9999)
		err = c.Compile(node.ThenBody)
		if err != nil {
			return err
		}

		// remove last OpPop to keep the last value of ThenBody in stack
		if c.lastOpIs(code.OpPop) {
			c.removeLastOp()
		}

		jumpPos := c.emit(code.OpJump, 9999)
		endOfThenBody := len(c.currentInstructions())
		c.replaceOperands(jumpNotTruethyPos, endOfThenBody)

		if node.ElseBody == nil {
			// if we don't have ElseBody, the end of ThenBody is the end for this If Expression
			c.emit(code.OpNull)
		} else {
			// if we have ElseBody we have to emit a OpJump as a part of the ThenBody
			// and let OpJumpNotTruethy jump over this OpJump to the start of the ElseBody
			err = c.Compile(node.ElseBody)
			if err != nil {
				return err
			}

			// same reason as remove last OpPop from ThenBody
			c.removeLastOp()
		}

		endOfElseBody := len(c.currentInstructions())
		c.replaceOperands(jumpPos, endOfElseBody)
	case *ast.LetStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		symbol := c.currentScope().localSymbolTable.Define(node.Name.Value)
		if symbol.Scope == GlobalScope {
			c.emit(code.OpSetGlobal, symbol.Index)
		} else {
			c.emit(code.OpSetLocal, symbol.Index)
		}
	case *ast.ArrayLiteral:
		var err error
		for _, e := range node.Elements {
			err = c.Compile(e)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpArray, len(node.Elements))
	case *ast.HashLiteral:
		var err error
		for k, v := range node.Pair {
			err = c.Compile(k)
			if err != nil {
				return err
			}
			err = c.Compile(v)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpHash, len(node.Pair))
	case *ast.Identifier:
		identifier := node.Value
		symbol, ok := c.currentScope().localSymbolTable.Resolve(identifier)
		if !ok {
			return fmt.Errorf("undefined variable %s", identifier)
		}

		if symbol.Scope == GlobalScope {
			c.emit(code.OpGetGlobal, symbol.Index)
		} else if symbol.Scope == BuiltinScope {
			c.emit(code.OpGetBuiltin, symbol.Index)
		} else {
			c.emit(code.OpGetLocal, symbol.Index)
		}
	case *ast.Integer:
		v := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(v))
	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	case *ast.String:
		v := &object.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(v))
	case *ast.ReturnStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		c.emit(code.OpReturnValue)
	case *ast.FunctionExpression:
		c.enterScope()

		for _, parameter := range node.Parameters {
			c.currentScope().localSymbolTable.Define(parameter.Value)
		}

		err := c.Compile(node.Body)
		if err != nil {
			return fmt.Errorf("compile function %s failed", node.Name)
		}

		if c.lastOpIs(code.OpPop) {
			c.replaceInstructions(c.currentScope().lastOpCodeStartPos, code.OpReturnValue)
		}

		if !c.lastOpIs(code.OpReturnValue) {
			c.emit(code.OpReturn)
		}

		scope := c.leaveScope()
		fn := &object.CompiledFunction{Instructions: scope.instructions,
			NumLocals:     scope.localSymbolTable.numDefinitions,
			NumParameters: len(node.Parameters)}
		c.emit(code.OpClosure, c.addConstant(fn), 0)
	case *ast.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}

		for _, e := range node.Arguments {
			err := c.Compile(e)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpCall, len(node.Arguments))
	default:
		return fmt.Errorf("unknown node type %T", node)
	}

	return nil
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{Instructions: c.currentInstructions(), Constants: c.constants}
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}
