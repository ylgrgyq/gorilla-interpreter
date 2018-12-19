package compiler

import (
	"ast"
	"code"
	"fmt"
	"object"
	"parser"
	"testing"
)

type compileTestCase struct {
	input              string
	expectInstructions []code.Instructions
	expectConstants    []object.Object
}

func parse(input string) (*ast.Program, error) {
	p := parser.New(input)
	program, err := p.ParseProgram()
	if err != nil {
		return nil, err
	}
	return program, nil
}

func testInstructions(expectIns code.Instructions, actualIns code.Instructions) error {
	if len(expectIns) != len(actualIns) {
		return fmt.Errorf("expect instructions length not equals to actual instructions length. want:%s got:%s",
			code.InstructionsString(expectIns), code.InstructionsString(actualIns))
	}

	for i, ins := range actualIns {
		if expectIns[i] != ins {
			return fmt.Errorf("compare instructions string failed. want:\n%s\ngot:\n%s",
				code.InstructionsString(expectIns), code.InstructionsString(actualIns))

		}
	}

	return nil
}

func testInteger(expect int64, actual object.Object) error {
	r, ok := actual.(*object.Integer)

	if !ok {
		return fmt.Errorf("compiled value not integer. got:%T (%+v)", actual, actual)
	}

	if expect != r.Value {
		return fmt.Errorf("compare integer failed. want:%d, got:%d", expect, r.Value)
	}

	return nil
}

func testConstants(expectConstants []object.Object, actualConstans []object.Object) error {
	if len(expectConstants) != len(actualConstans) {
		return fmt.Errorf("compare constants length failed. want:%d got:%d",
			len(expectConstants), len(actualConstans))
	}

	index := 0
	for index < len(actualConstans) {
		expect := expectConstants[index]
		actual := actualConstans[index]

		switch expect := expect.(type) {
		case *object.Integer:
			return testInteger(expect.Value, actual)
		}
		index++
	}

	return nil
}

func runTests(t *testing.T, tests []compileTestCase) {
	t.Helper()

	for _, test := range tests {
		program, err := parse(test.input)
		if err != nil {
			t.Errorf("parse input %s failed %s", test.input, err)
		}

		c := New()
		err = c.Compile(program)
		if err != nil {
			t.Errorf("compile input %s failed %s", test.input, err)
		}

		codes := c.Bytecode()
		actualIns := codes.Instructions
		cons := codes.Constants
		expectIns := code.FlattenInstructions(test.expectInstructions)

		err = testInstructions(expectIns, actualIns)
		if err != nil {
			t.Errorf("Error for input: %s. %s", test.input, err)
		}

		err = testConstants(test.expectConstants, cons)
		if err != nil {
			t.Errorf("Error for input: %s. %s", test.input, err)
		}
	}
}
func TestCompileIntegerArithmetic(t *testing.T) {
	tests := []compileTestCase{
		{"3", []code.Instructions{code.Make(code.OpConstant, 0)}, []object.Object{&object.Integer{Value: 3}}},
		{"1 + 2",
			[]code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd)},
			[]object.Object{
				&object.Integer{Value: 1},
				&object.Integer{Value: 23}}},
	}

	runTests(t, tests)
}
