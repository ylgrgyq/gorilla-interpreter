package compiler

import (
	"ast"
	"code"
	"fmt"
	"object"
)

type Compiler struct {
	instructions code.Instructions
	constants    []object.Object

	lastOpCodeStartPos       int
	secondLastOpCodeStartPos int
}

func New() *Compiler {
	return &Compiler{instructions: []byte{}, constants: []object.Object{}}
}

func (c *Compiler) addConstant(value object.Object) int {
	c.constants = append(c.constants, value)
	return len(c.constants) - 1
}

func (c *Compiler) updateLastOpCodeStartPos(lastOpCodeStartPos int) {
	secondLast := c.lastOpCodeStartPos
	c.lastOpCodeStartPos = lastOpCodeStartPos
	c.secondLastOpCodeStartPos = secondLast
}

func (c *Compiler) emit(op code.OpCode, operands ...int) int {
	ins := code.Make(op, operands...)
	startPos := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	c.updateLastOpCodeStartPos(startPos)
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
	codeToReplace := c.instructions[opCodeStartPos]
	newIns := code.Make(code.OpCode(codeToReplace), newOperands...)
	for i, ins := range newIns {
		c.instructions[i+opCodeStartPos] = ins
	}
}

func (c *Compiler) removeLastOpPop() {
	lastOpCode := code.OpCode(c.instructions[c.lastOpCodeStartPos])
	if lastOpCode == code.OpPop {
		c.instructions = c.instructions[:c.lastOpCodeStartPos]
		c.lastOpCodeStartPos = c.secondLastOpCodeStartPos
		c.secondLastOpCodeStartPos = -1
	}
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
		c.removeLastOpPop()

		jumpPos := c.emit(code.OpJump, 9999)
		endOfThenBody := len(c.instructions)
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
			c.removeLastOpPop()
		}

		endOfElseBody := len(c.instructions)
		c.replaceOperands(jumpPos, endOfElseBody)

	case *ast.Integer:
		v := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(v))
	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	default:
		return fmt.Errorf("unknown node type %T", node)
	}

	return nil
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{Instructions: c.instructions, Constants: c.constants}
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}
