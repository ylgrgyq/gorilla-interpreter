package vm

import (
	"code"
	"object"
)

type Frame struct {
	fn *object.CompiledFunction
	ip int
	// base pointer points to the CompileFunction for this Frame
	basePointer int
}

func NewFrame(fn *object.CompiledFunction, basePointer int) *Frame {
	return &Frame{
		fn: fn,
		ip: 0,
		basePointer: basePointer,
	}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
