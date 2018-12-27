package vm

import (
	"code"
	"compiler"
	"fmt"
	"object"
)

const StackSize = 2048
const GlobalSize = 65535

type VM struct {
	instructions code.Instructions
	constants    []object.Object

	stack []object.Object
	sp    int

	lastPop object.Object

	globals []object.Object
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		stack:        make([]object.Object, StackSize),
		sp:           -1,
		globals:      make([]object.Object, GlobalSize),
	}
}

func NewWithGlobals(bytecode *compiler.Bytecode, globals []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = globals
	return vm
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

func (v *VM) executeBinaryOperatorOnInteger(op code.OpCode, l int64, r int64) (object.Object, error) {
	var result object.Object
	switch op {
	case code.OpAdd:
		result = &object.Integer{Value: l + r}
	case code.OpSubtraction:
		result = &object.Integer{Value: l - r}
	case code.OpMultiply:
		result = &object.Integer{Value: l * r}
	case code.OpDivide:
		result = &object.Integer{Value: l / r}
	case code.OpEqual:
		result = &object.Boolean{Value: l == r}
	case code.OpNotEqual:
		result = &object.Boolean{Value: l != r}
	case code.OpGreaterEqual:
		result = &object.Boolean{Value: l >= r}
	case code.OpGreaterThan:
		result = &object.Boolean{Value: l > r}
	default:
		return nil, fmt.Errorf("unsupportted operator on integer: %d", op)
	}

	return result, nil
}

func (v *VM) executeBinaryOperatorOnBoolean(op code.OpCode, l bool, r bool) (object.Object, error) {
	var result object.Object
	switch op {
	case code.OpEqual:
		result = &object.Boolean{Value: l == r}
	case code.OpNotEqual:
		result = &object.Boolean{Value: l != r}
	default:
		return nil, fmt.Errorf("unsupportted operator on boolean: %d", op)
	}

	return result, nil
}

func (v *VM) executeBinaryOperatorOnString(op code.OpCode, l string, r string) (object.Object, error) {
	var result object.Object
	switch op {
	case code.OpEqual:
		result = &object.Boolean{Value: l == r}
	case code.OpNotEqual:
		result = &object.Boolean{Value: l != r}
	case code.OpAdd:
		result = &object.String{Value: l + r}
	default:
		return nil, fmt.Errorf("unsupportted operator on string: %d", op)
	}

	return result, nil
}

func (v *VM) executeBinaryOperator(op code.OpCode) error {
	right := v.popStack()
	left := v.popStack()

	var result object.Object
	var err error
	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		l := left.(*object.Integer).Value
		r := right.(*object.Integer).Value
		result, err = v.executeBinaryOperatorOnInteger(op, l, r)
		if err != nil {
			return err
		}
	} else if left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ {
		l := left.(*object.Boolean).Value
		r := right.(*object.Boolean).Value
		result, err = v.executeBinaryOperatorOnBoolean(op, l, r)
		if err != nil {
			return err
		}
	} else if left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ {
		l := left.(*object.String).Value
		r := right.(*object.String).Value
		result, err = v.executeBinaryOperatorOnString(op, l, r)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unsupportted binary operator %d with %T and %T as operands", op, left, right)
	}

	err = v.pushStack(result)
	return err
}

func (v *VM) executeBangOperator() error {
	val := v.popStack()
	if val == nil {
		return fmt.Errorf("bang operator need one operand")
	}

	switch val {
	case object.TRUE:
		return v.pushStack(object.FALSE)
	case object.FALSE, object.NULL:
		return v.pushStack(object.TRUE)
	default:
		return v.pushStack(object.FALSE)
	}
}

func (v *VM) executeMinusOperator() error {
	val := v.popStack()

	if val == nil {
		return fmt.Errorf("minus operator need one operand")
	}

	if val.Type() == object.INTEGER_OBJ {
		intV := val.(*object.Integer).Value
		return v.pushStack(&object.Integer{Value: -intV})
	}

	return fmt.Errorf("unsupportted minus operator on %T value", val)
}

func isTruethy(obj object.Object) bool {
	if obj == nil {
		return false
	}

	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Internal_Null:
		return false
	default:
		return true
	}
}

func (v *VM) Run() error {
	var err error
	for ip := 0; ip < len(v.instructions); ip++ {
		c := code.OpCode(v.instructions[ip])

		switch c {
		case code.OpConstant:
			index := code.ReadUint16(v.instructions[ip+1:])
			ip += 2

			err = v.pushStack(v.constants[index])
		case code.OpSetGlobal:
			index := code.ReadUint16(v.instructions[ip+1:])
			ip += 2
			globalV := v.popStack()
			v.globals[index] = globalV
		case code.OpGetGlobal:
			index := code.ReadUint16(v.instructions[ip+1:])
			ip += 2
			globalV := v.globals[index]
			v.pushStack(globalV)
		case code.OpNull:
			err = v.pushStack(object.NULL)
		case code.OpBang:
			err = v.executeBangOperator()
		case code.OpMinus:
			err = v.executeMinusOperator()
		case code.OpAdd, code.OpSubtraction, code.OpMultiply, code.OpDivide,
			code.OpEqual, code.OpNotEqual, code.OpGreaterEqual, code.OpGreaterThan:
			err = v.executeBinaryOperator(c)
		case code.OpTrue:
			err = v.pushStack(object.TRUE)
		case code.OpFalse:
			err = v.pushStack(object.FALSE)
		case code.OpPop:
			v.popStack()
		case code.OpJumptNotTruethy:
			targetPos := int(code.ReadUint16(v.instructions[ip+1:]))
			ip += 2

			conditionVal := v.popStack()
			if !isTruethy(conditionVal) {
				ip = targetPos - 1
			}
		case code.OpJump:
			targetPos := int(code.ReadUint16(v.instructions[ip+1:]))
			ip = targetPos - 1
		}

		if err != nil {
			return err
		}
	}

	return nil
}
