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
	case *ast.InfixExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}
	case *ast.Integer:
		v := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(v))
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
