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

func testBooleanObject(expected bool, actual object.Object) error {
	r, ok := actual.(*object.Boolean)
	if !ok {
		return fmt.Errorf("object is not object.Boolean. got=%T (%+v)", actual, actual)
	}

	if expected != r.Value {
		return fmt.Errorf("assert failed. want=%t, got=%t", expected, r.Value)
	}

	return nil
}

func testNilObject(actual object.Object) error {
	_, ok := actual.(*object.Internal_Null)
	if !ok {
		return fmt.Errorf("object is not NULL. got=%T (%+v)", actual, actual)
	}

	return nil
}

func testExpectedObject(t *testing.T, input string, expcet interface{}, actual object.Object) {
	t.Helper()

	switch expected := expcet.(type) {
	case int:
		err := testIntegerObject(int64(expected), actual)
		if err != nil {
			t.Errorf("test integer object failed for input: %s. %s", input, err)
		}
	case bool:
		err := testBooleanObject(expected, actual)
		if err != nil {
			t.Errorf("test boolean object failed for input: %s. %s", input, err)
		}
	case nil:
		err := testNilObject(actual)
		if err != nil {
			t.Errorf("test boolean object failed for input: %s. %s", input, err)
		}
	default:
		t.Errorf("unknown type %T for input %s", expcet, input)
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

		testExpectedObject(t, test.input, test.expect, v.StackLastTop())
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
		{"-1 + 2", 1},
	}
	runTests(t, tests)
}

func TestComparation(t *testing.T) {
	tests := []vmTestCase{
		{"true", true},
		{"false", false},
		{"!false == true", true},
		{"true == true", true},
		{"true != false", true},
		{"1 + 2 < 5", true},
		{"1 + 2 < 1", false},
		{"1 + 2 <= 2 + 1", true},
		{"1 + 2 >= 2 + 1", true},
		{"1 + 4 >= 8 + 1", false},
	}
	runTests(t, tests)
}

func TestIfExpression(t *testing.T) {
	tests := []vmTestCase{
		{"if (true) {100}", 100},
		{"if (false) {100} else {50}", 50},
		{"if (1 < 10) {99} else {11}", 99},
		{"if (102 >= 1000) {99} else {11}", 11},
		{"if (false) {100}", nil},
	}
	runTests(t, tests)
}

func TestLetStatement(t *testing.T) {
	tests := []vmTestCase{
		{"let a = 1; a;", 1},
		{"let a = 1; let b = a;  b;", 1},
	}

	runTests(t, tests)
}
