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
	OpAdd
	OpMinus
	OpMultiply
	OpDivide
	OpEqual
	OpNotEqual
	OpGreaterThan
	OpGreaterEqual
	OpPop
)

type Definition struct {
	Name         string
	OperandWiths []int
}

var definitionMap = map[OpCode]*Definition{
	OpConstant:     &Definition{"OpConstant", []int{2}},
	OpTrue:         &Definition{"OpTrue", []int{}},
	OpFalse:        &Definition{"OpFalse", []int{}},
	OpAdd:          &Definition{"OpAdd", []int{}},
	OpMinus:        &Definition{"OpMinus", []int{}},
	OpMultiply:     &Definition{"OpMultiply", []int{}},
	OpDivide:       &Definition{"OpDivide", []int{}},
	OpEqual:        &Definition{"OpEqual", []int{}},
	OpNotEqual:     &Definition{"OpNotEqual", []int{}},
	OpGreaterThan:  &Definition{"OpGreaterThan", []int{}},
	OpGreaterEqual: &Definition{"OpGreaterEqual", []int{}},
	OpPop:          &Definition{"OpPop", []int{}},
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

func ReadOperand(def *Definition, ins Instructions) ([]int, int) {
	ret := make([]int, len(def.OperandWiths))

	offset := 0
	for i, width := range def.OperandWiths {
		switch width {
		case 2:
			ret[i] = int(ReadUint16(ins[offset:]))
		}

		offset += width
	}

	return ret, offset
}

func fmtInstruction(def *Definition, operands ...int) string {
	operandCount := len(def.OperandWiths)

	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	}

	return fmt.Sprintf("unhandled operand count for code %s", def.Name)
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
