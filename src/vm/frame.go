package vm

import (
	"code"
	"object"
)

type Frame struct {
	clo         *object.Closure
	ip          int
	basePointer int
}

func NewFrame(clo *object.Closure, basePointer int) *Frame {
	return &Frame{
		clo:         clo,
		ip:          0,
		basePointer: basePointer,
	}
}

func (f *Frame) Instructions() code.Instructions {
	return f.clo.Fn.Instructions
}
