package vm

import (
	"code"
	"compiler"
	"fmt"
	"object"
)

const StackSize = 2048

type VM struct {
	instructions code.Instructions
	constants    []object.Object

	stack []object.Object
	sp    int

	lastPop object.Object
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		stack:        make([]object.Object, StackSize),
		sp:           -1,
	}
}

func (v *VM) pushStack(o object.Object) error {
	if v.sp >= len(v.stack) {
		return fmt.Errorf("Stack full")
	}

	v.sp++
	v.stack[v.sp] = o
	return nil
}

func (v *VM) popStack() object.Object {
	if v.sp < 0 {
		return nil
	}

	o := v.stack[v.sp]
	v.lastPop = o
	v.sp--
	return o
}

func (v *VM) StackTop() object.Object {
	if v.sp < 0 {
		return nil
	}

	return v.stack[v.sp]
}

func (v *VM) StackLastTop() object.Object {
	return v.lastPop
}

func (v *VM) Run() error {
	for ip := 0; ip < len(v.instructions); ip++ {
		c := code.OpCode(v.instructions[ip])

		switch c {
		case code.OpConstant:
			index := code.ReadUint16(v.instructions[ip+1:])
			ip += 2

			err := v.pushStack(v.constants[index])
			if err != nil {
				return err
			}
		case code.OpAdd:
			left := v.popStack()
			right := v.popStack()

			l := left.(*object.Integer).Value
			r := right.(*object.Integer).Value

			err := v.pushStack(&object.Integer{Value: l + r})
			if err != nil {
				return err
			}
		case code.OpPop:
			v.popStack()
		}
	}

	return nil
}
