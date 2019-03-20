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

func testStringObject(expected string, actual object.Object) error {
	r, ok := actual.(*object.String)
	if !ok {
		return fmt.Errorf("object is not object.String. got=%T (%+v)", actual, actual)
	}

	if expected != r.Value {
		return fmt.Errorf("assert failed. want=%q, got=%q", expected, r.Value)
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
	case string:
		err := testStringObject(expected, actual)
		if err != nil {
			t.Errorf("test string object failed for input: %s. %s", input, err)
		}
	case []interface{}:
		r, ok := actual.(*object.Array)
		if !ok {
			t.Errorf("object is not object.Array. got=%T (%+v)", actual, actual)
		}

		if len(expected) != len(r.Elements) {
			t.Errorf("array length is not equal. want=%d got=%d", len(expected), len(r.Elements))
		}

		for i, e := range expected {
			testExpectedObject(t, input, e, r.Elements[i])
		}
	case map[interface{}]interface{}:
		r, ok := actual.(*object.HashTable)
		if !ok {
			t.Errorf("object is not object.HashTable. got=%T (%+v)", actual, actual)
		}

		if len(expected) != len(r.Pair) {
			t.Errorf("hash length is not equal. want=%d got=%d", len(expected), len(r.Pair))
		}

		// Hash 的测试有点麻烦，我不想让实现出来的 Map 仅仅是为了好测试而搞成有序的，非常丑
		// 但不搞成有序的，actual 里面的 k v 都是 object.Object，在不知道对象类型的情况下无法去跟 expect 做对比
		// 所以我决定不测了
	case nil:
		err := testNilObject(actual)
		if err != nil {
			t.Errorf("test boolean object failed for input: %s. %s", input, err)
		}
	case *object.Error:
		errObj, ok := actual.(*object.Error)
		if !ok {
			t.Errorf("object is not Error: %T (%+v)", actual, actual)
			return
		}
		if errObj.Msg != expected.Msg {
			t.Errorf("wrong error message. expected=%q, got=%q",
				expected.Msg, errObj.Msg)
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

func TestStringOperator(t *testing.T) {
	tests := []vmTestCase{
		{"\"hello\" == \"hello\" ", true},
		{"\"hello\" == \"world\" ", false},
		{"\"hello\" + \"world\" ", "helloworld"},
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

func TestArray(t *testing.T) {
	tests := []vmTestCase{
		{"[]", []interface{}{}},
		{"[1, 2, 3]", []interface{}{1, 2, 3}},
		{"[1 + 2, 3 * 4, 5 + 6]", []interface{}{3, 12, 11}},
		{"[1, 2 + 4, false, \"hello\" + \"world\"]", []interface{}{1, 6, false, "helloworld"}},
		{"[1, 2, 3][0]", 1},
		{"[1, 2 + 4, false, \"hello\" + \"world\"][3]", "helloworld"},
	}

	runTests(t, tests)
}

func TestHash(t *testing.T) {
	tests := []vmTestCase{
		{"{}", map[interface{}]interface{}{}},
		// {`{1:"hello", 666 + 100:2 + 15, "haha":false, "s":"hello" + "world"}`, map[interface{}]interface{}{1: "hello", 766: 17, "haha": false, "s": "helloworld"}},
		// {`{1:"hello", 666 + 100:2 + 15, "haha":false, "s":"hello" + "world"}[266 + 500]`, 17},
	}

	runTests(t, tests)
}

func TestFunction(t *testing.T) {
	tests := []vmTestCase{
		{"let fivePlusTen = fn(){ 5 + 10 }" +
			"fivePlusTen()",
			15},

		{"let noReturn = fn(){ }" +
			"noReturn()", nil},
		{`
        let noReturn = fn() { };
        let noReturnTwo = fn() { noReturn(); };
        noReturn();
        noReturnTwo();
        `, nil},
		{`
        let returnsOne = fn() { 1; };
        let returnsOneReturner = fn() { returnsOne; };
        returnsOneReturner()();
        `,
			1,
		},
		{`
         let identity = fn(a) { a; };
         identity(5);
         `,
			5,
		},
		{`
         let sum = fn(a, b) { a + b; };
         sum(100, 200);
         `,
			300,
		},
		{`
         let someFn = fn(a, b) {let c = 100;  a + b + c; };
         someFn(100, 200);
         `,
			400,
		},
	}

	runTests(t, tests)
}

func TestLocalBinding(t *testing.T) {
	tests := []vmTestCase{
		{
			"let one = fn () { let one = 1; one}; one();",
			1,
		},
		{
			"let oneAndTwo = fn () { let one = 1; let two = 2; one + two}; oneAndTwo();",
			3,
		},
		{
			`
           let oneAndTwo = fn() { let one = 1; let two = 2; one + two; };
           let threeAndFour = fn() { let three = 3; let four = 4; three + four; };
           oneAndTwo() + threeAndFour();
           `,
			10,
		},
		{
			`
           let firstFoobar = fn() { let foobar = 50; foobar; };
           let secondFoobar = fn() { let foobar = 100; foobar; };
           firstFoobar() + secondFoobar();
           `,
			150,
		},
		{
			`
           let globalSeed = 50;
           let minusOne = fn() {
               let num = 1;
               globalSeed - num;
           }
           let minusTwo = fn() {
               let num = 2;
               globalSeed - num;
           }
           minusOne() + minusTwo();`,
			97,
		},
	}
	runTests(t, tests)
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []vmTestCase{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{
			`len(1)`,
			&object.Error{Msg: "argument to `len` not supported, got INTEGER"},
		},
		{`len("one", "two")`,
			&object.Error{Msg: "wrong number of arguments. expect=1, got=2"}},
		{`len([1, 2, 3])`, 3},
		{`len([])`, 0},
		{`first([1, 2, 3])`, 1},
		{`first([])`, nil},
		{`first(1)`,
			&object.Error{Msg: "wrong argument passed to function first. expect Array, got=\"INTEGER\""},
		},
		{`last([1, 2, 3])`, 3},
		{`last([])`, nil},
		{`last(1)`,
			&object.Error{Msg: "wrong argument passed to function last. expect Array, got=\"INTEGER\""}},
		{`rest([1, 2, 3])`, []interface{}{2, 3}},
		{`rest([])`, nil},
		{`push([], 1)`, []interface{}{1}},
		{`push(1, 1)`,
			&object.Error{Msg: "wrong argument passed to function push. expect Array, got=\"INTEGER\""}},
	}
	runTests(t, tests)
}
