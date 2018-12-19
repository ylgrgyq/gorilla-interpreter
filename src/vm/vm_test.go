package vm

import (
	"ast"
	"compiler"
	"fmt"
	"object"
	"parser"
	"testing"
)

type vmTestCase struct {
	input  string
	expect interface{}
}

func parse(input string) (*ast.Program, error) {
	p := parser.New(input)
	program, err := p.ParseProgram()
	if err != nil {
		return nil, err
	}
	return program, nil
}

func testIntegerObject(expected int64, actual object.Object) error {
	r, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not object.Integer. got=%T (%+v)", actual, actual)
	}

	if expected != r.Value {
		return fmt.Errorf("assert failed. want=%d, got=%d", expected, r.Value)
	}

	return nil
}

func testExpectedObject(t *testing.T, expcet interface{}, actual object.Object) {
	t.Helper()

	switch expected := expcet.(type) {
	case int:
		err := testIntegerObject(int64(expected), actual)
		if err != nil {
			t.Errorf("test integer object failed: %s", err)
		}
	default:
		t.Errorf("unknown type %T", expcet)
	}
}

func runTests(t *testing.T, tests []vmTestCase) {
	t.Helper()
	for _, test := range tests {
		program, err := parse(test.input)
		if err != nil {
			t.Fatalf("parse program failed. %s", err)
		}

		c := compiler.New()
		err = c.Compile(program)
		if err != nil {
			t.Fatalf("compile program for input: %q failed. error is: %q", test.input, err)
		}

		v := New(c.Bytecode())
		err = v.Run()
		if err != nil {
			t.Fatalf("run program for input: %q failed. error is: %q", test.input, err)
		}

		testExpectedObject(t, test.expect, v.StackLastTop())
		if v.StackTop() != nil {
			t.Fatalf("left %s in stack.", v.StackTop().Inspect())
		}
	}
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"2", 2},
		{"1 + 2", 3},
		{"4 + 4", 8},
		{"4 - 4 * 15 / 2", -26},
	}
	runTests(t, tests)
}
