package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instructions []byte
type OpCode byte

const (
	OpConstant OpCode = iota
	OpTrue
	OpFalse
	OpArray
	OpHash
	OpAdd
	OpSubtraction
	OpMultiply
	OpDivide
	OpMinus
	OpEqual
	OpNotEqual
	OpGreaterThan
	OpGreaterEqual
	OpBang
	OpIndex
	OpJumptNotTruethy
	OpJump
	OpGetGlobal
	OpSetGlobal
	OpSetLocal
	OpGetLocal
	OpPop
	OpNull
	OpCall
	OpReturnValue
	OpReturn
	OpGetBuiltin
	OpClosure
	OpGetFree
	OpCurrentClosure
)

type Definition struct {
	Name         string
	OperandWiths []int
}

var definitionMap = map[OpCode]*Definition{
	OpConstant:        &Definition{"OpConstant", []int{2}},
	OpTrue:            &Definition{"OpTrue", []int{}},
	OpFalse:           &Definition{"OpFalse", []int{}},
	OpArray:           &Definition{"OpArray", []int{2}},
	OpHash:            &Definition{"OpHash", []int{2}},
	OpAdd:             &Definition{"OpAdd", []int{}},
	OpSubtraction:     &Definition{"OpSubtraction", []int{}},
	OpMultiply:        &Definition{"OpMultiply", []int{}},
	OpDivide:          &Definition{"OpDivide", []int{}},
	OpMinus:           &Definition{"OpMinus", []int{}},
	OpEqual:           &Definition{"OpEqual", []int{}},
	OpNotEqual:        &Definition{"OpNotEqual", []int{}},
	OpGreaterThan:     &Definition{"OpGreaterThan", []int{}},
	OpGreaterEqual:    &Definition{"OpGreaterEqual", []int{}},
	OpBang:            &Definition{"OpBang", []int{}},
	OpIndex:           &Definition{"OpIndex", []int{}},
	OpJumptNotTruethy: &Definition{"OpJumpNotTruethy", []int{2}},
	OpJump:            &Definition{"OpJump", []int{2}},
	OpGetGlobal:       &Definition{"OpGetGlobal", []int{2}},
	OpSetGlobal:       &Definition{"OpSetGlobal", []int{2}},
	OpGetLocal:        &Definition{"OpGetLocal", []int{1}},
	OpSetLocal:        &Definition{"OpSetLocal", []int{1}},
	OpPop:             &Definition{"OpPop", []int{}},
	OpNull:            &Definition{"OpNull", []int{}},
	OpCall:            &Definition{"OpCall", []int{1}},
	OpReturnValue:     &Definition{"OpReturnValue", []int{}},
	OpReturn:          &Definition{"OpReturn", []int{}},
	OpGetBuiltin:      &Definition{"OpGetBuiltin", []int{1}},
	OpClosure:         &Definition{"OpClosure", []int{2, 1}},
	OpGetFree:         &Definition{"OpGetFree", []int{1}},
	OpCurrentClosure:  &Definition{"OpCurrentClosure", []int{}},
}

func Lookup(code OpCode) (*Definition, error) {
	def, ok := definitionMap[code]
	if !ok {
		return nil, fmt.Errorf("can not find definition for code %d", code)
	}
	return def, nil
}

func Make(op OpCode, operands ...int) []byte {
	def, ok := definitionMap[op]
	if !ok {
		return []byte{}
	}

	if len(operands) != len(def.OperandWiths) {
		return []byte{}
	}

	// at least one byte to store OpCode for instructions
	length := 1
	for _, width := range def.OperandWiths {
		length += width
	}

	instructions := make([]byte, length)
	instructions[0] = byte(op)

	offset := 1
	for i, operand := range operands {
		expectWidth := def.OperandWiths[i]
		switch expectWidth {
		case 1:
			instructions[offset] = byte(operand)
		case 2:
			binary.BigEndian.PutUint16(instructions[offset:], uint16(operand))
		}
		offset += expectWidth
	}

	return instructions
}

func ReadUint16(bs Instructions) uint16 {
	return binary.BigEndian.Uint16(bs)
}

func ReadUint8(bs Instructions) uint8 {
	return uint8(bs[0])
}

func ReadOperand(def *Definition, ins Instructions) ([]int, int) {
	ret := make([]int, len(def.OperandWiths))

	offset := 0
	for i, width := range def.OperandWiths {
		switch width {
		case 2:
			ret[i] = int(ReadUint16(ins[offset:]))
		case 1:
			ret[i] = int(ReadUint8(ins[offset:]))
		}

		offset += width
	}

	return ret, offset
}

func fmtInstruction(def *Definition, operands ...int) string {
	operandCount := len(def.OperandWiths)

	ret := fmt.Sprintf("%s", def.Name)
	for i, op := range operands {
		if i >= operandCount {
			break
		}
		ret += fmt.Sprintf(" %d", op)
	}

	return ret
}

func InstructionsString(ins Instructions) string {
	var out bytes.Buffer

	cursor := 0
	for cursor < len(ins) {
		code := OpCode(ins[cursor])
		def, err := Lookup(code)
		if err != nil {
			fmt.Fprintf(&out, "Error: %s\n", err)
			break
		}

		operands, read := ReadOperand(def, ins[cursor+1:])
		if len(operands) != len(def.OperandWiths) {
			fmt.Fprintf(&out, "Error: no enough operands for code:%s. want:%d, got:%d", def.Name, len(def.OperandWiths), len(operands))
			break
		}

		fmt.Fprintf(&out, "%04d %s\n", cursor, fmtInstruction(def, operands...))
		cursor += read + 1
	}

	return out.String()
}

func FlattenInstructions(ins []Instructions) Instructions {
	out := Instructions{}
	for _, i := range ins {
		out = append(out, i...)
	}

	return out
}
