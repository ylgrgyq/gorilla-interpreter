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
}

func New() *Compiler {
	return &Compiler{instructions: []byte{}, constants: []object.Object{}}
}

func (c *Compiler) addConstant(value object.Object) int {
	c.constants = append(c.constants, value)
	return len(c.constants) - 1
}

func (c *Compiler) emit(op code.OpCode, operands ...int) int {
	ins := code.Make(op, operands...)
	startPos := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
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
		c.emit(code.OpMinus)
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

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
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

	case *ast.InfixExpression:
		err := c.compileInfixExpression(node)
		if err != nil {
			return err
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
