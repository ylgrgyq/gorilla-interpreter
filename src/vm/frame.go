package vm

import (
	"code"
	"object"
)

type Frame struct {
	fn *object.CompiledFunction
	ip int
	locals []object.Object
}

func NewFrame(fn *object.CompiledFunction) *Frame {
	return &Frame{
		fn: fn,
		ip: 0,
		locals: make([]object.Object, fn.NumLocals),
	}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
