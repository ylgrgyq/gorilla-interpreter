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
	expectConstants    []interface{}
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
		return fmt.Errorf("expect instructions length not equals to actual instructions length. want:\n%s\ngot:\n%s",
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

func testBoolean(expect bool, actual object.Object) error {
	r, ok := actual.(*object.Boolean)

	if !ok {
		return fmt.Errorf("compiled value not boolean. got:%T (%+v)", actual, actual)
	}

	if expect != r.Value {
		return fmt.Errorf("compare boolean failed. want:%t, got:%t", expect, r.Value)
	}

	return nil
}

func testString(expect string, actual object.Object) error {
	r, ok := actual.(*object.String)

	if !ok {
		return fmt.Errorf("compiled value not string. got:%T (%+v)", actual, actual)
	}

	if expect != r.Value {
		return fmt.Errorf("compare string failed. want:%q, got:%q", expect, r.Value)
	}

	return nil
}

func testConstants(expectConstants []interface{}, actualConstans []object.Object) error {
	if len(expectConstants) != len(actualConstans) {
		return fmt.Errorf("compare constants length failed. want:%d got:%d",
			len(expectConstants), len(actualConstans))
	}

	index := 0
	for index < len(actualConstans) {
		expect := expectConstants[index]
		actual := actualConstans[index]

		switch expect := expect.(type) {
		case int:
			err := testInteger(int64(expect), actual)
			if err != nil {
				return err
			}
		case bool:
			err := testBoolean(expect, actual)
			if err != nil {
				return err
			}
		case string:
			err := testString(expect, actual)
			if err != nil {
				return err
			}
		case []code.Instructions:
			fn, ok := actual.(*object.CompiledFunction)
			if !ok {
				return fmt.Errorf("constant %d - not a function: %T", index, actual)
			}
			expectIns := code.FlattenInstructions(expect)
			err := testInstructions(expectIns, fn.Instructions)
			if err != nil {
				return fmt.Errorf("constant %d - testInstructions failed: %s", index, err)
			}
		case object.CompiledFunction:
			fn, ok := actual.(*object.CompiledFunction)
			if !ok {
				return fmt.Errorf("constant %d - not a function: %T", index, actual)
			}

			err := testInstructions(expect.Instructions, fn.Instructions)
			if err != nil {
				return fmt.Errorf("constant %d - testInstructions failed: %s", index, err)
			}
			if expect.NumLocals != fn.NumLocals {
				return fmt.Errorf("constant %d - test NumLocals failed. want: %d, got: %d",
					index, expect.NumLocals, fn.NumLocals)
			}
		default:
			return fmt.Errorf("constant %d - unknown expect type %T", index, expect)
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
			t.Errorf("Compile instructions error for input: %s. %s", test.input, err)
		}

		err = testConstants(test.expectConstants, cons)
		if err != nil {
			t.Errorf("Compile constant error for input: %s. %s", test.input, err)
		}
	}
}
func TestCompileIntegerArithmetic(t *testing.T) {
	tests := []compileTestCase{
		{"3", []code.Instructions{code.Make(code.OpConstant, 0),
			code.Make(code.OpPop)},
			[]interface{}{3}},
		{"-3", []code.Instructions{
			code.Make(code.OpConstant, 0),
			code.Make(code.OpMinus),
			code.Make(code.OpPop)},
			[]interface{}{3}},
		{"-1 + 2",
			[]code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpMinus),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop)},
			[]interface{}{
				1,
				2}},
		{"(1 - 2 - 3) / 100",
			[]code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSubtraction),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpSubtraction),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpDivide),
				code.Make(code.OpPop)},
			[]interface{}{
				1,
				2,
				3,
				100}},
		{"(1 - 2 * 3) / 100",
			[]code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpMultiply),
				code.Make(code.OpSubtraction),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpDivide),
				code.Make(code.OpPop)},
			[]interface{}{
				1,
				2,
				3,
				100,
			}},
		{"4 - 4 * 15 / 2",
			[]code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpMultiply),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpDivide),
				code.Make(code.OpSubtraction),
				code.Make(code.OpPop)},
			[]interface{}{
				4,
				4,
				15,
				2,
			}},
	}

	runTests(t, tests)
}

func TestCompileBoolean(t *testing.T) {
	tests := []compileTestCase{
		{"true", []code.Instructions{code.Make(code.OpTrue),
			code.Make(code.OpPop)},
			[]interface{}{}},
		{"false", []code.Instructions{code.Make(code.OpFalse),
			code.Make(code.OpPop)},
			[]interface{}{}},
		{"!false == true", []code.Instructions{
			code.Make(code.OpFalse),
			code.Make(code.OpBang),
			code.Make(code.OpTrue),
			code.Make(code.OpEqual),
			code.Make(code.OpPop)},
			[]interface{}{}},
		{"((15 + 221) == 236) == (false != (68 < 103))", []code.Instructions{
			code.Make(code.OpConstant, 0),
			code.Make(code.OpConstant, 1),
			code.Make(code.OpAdd),
			code.Make(code.OpConstant, 2),
			code.Make(code.OpEqual),

			code.Make(code.OpFalse),
			code.Make(code.OpConstant, 3),
			code.Make(code.OpConstant, 4),
			code.Make(code.OpGreaterThan),
			code.Make(code.OpNotEqual),
			code.Make(code.OpEqual),
			code.Make(code.OpPop)},
			[]interface{}{
				15,
				221,
				236,
				103,
				68,
			}},

		{"(68 - 25) <= 236", []code.Instructions{
			code.Make(code.OpConstant, 0),
			code.Make(code.OpConstant, 1),
			code.Make(code.OpConstant, 2),
			code.Make(code.OpSubtraction),
			code.Make(code.OpGreaterEqual),
			code.Make(code.OpPop)},
			[]interface{}{
				236,
				68,
				25,
			}},
		{"(68 - 25) > 21", []code.Instructions{
			code.Make(code.OpConstant, 0),
			code.Make(code.OpConstant, 1),
			code.Make(code.OpSubtraction),
			code.Make(code.OpConstant, 2),
			code.Make(code.OpGreaterThan),
			code.Make(code.OpPop)},
			[]interface{}{
				68,
				25,
				21,
			}},
		{"(68 - 25) >= 21", []code.Instructions{
			code.Make(code.OpConstant, 0),
			code.Make(code.OpConstant, 1),
			code.Make(code.OpSubtraction),
			code.Make(code.OpConstant, 2),
			code.Make(code.OpGreaterEqual),
			code.Make(code.OpPop)},
			[]interface{}{
				68,
				25,
				21,
			}},
	}

	runTests(t, tests)
}

func TestConditional(t *testing.T) {
	tests := []compileTestCase{
		{"if (true) {100}; 9999",
			[]code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpJumptNotTruethy, 10),
				code.Make(code.OpConstant, 0),
				code.Make(code.OpJump, 11),
				code.Make(code.OpNull),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
			[]interface{}{
				100,
				9999,
			}},
		{"if (false) {100}; 9999",
			[]code.Instructions{
				code.Make(code.OpFalse),
				code.Make(code.OpJumptNotTruethy, 10),
				code.Make(code.OpConstant, 0),
				code.Make(code.OpJump, 11),
				code.Make(code.OpNull),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
			[]interface{}{
				100,
				9999,
			}},
		{"if (false) {100} else {50}; 9999",
			[]code.Instructions{
				code.Make(code.OpFalse),
				code.Make(code.OpJumptNotTruethy, 10),
				code.Make(code.OpConstant, 0),
				code.Make(code.OpJump, 13),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpPop),
			},
			[]interface{}{
				100,
				50,
				9999,
			}},
	}

	runTests(t, tests)
}

func TestGetSetGlobal(t *testing.T) {
	tests := []compileTestCase{
		{`let a = 1;
		  let b = 2;`,
			[]code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSetGlobal, 1),
			},
			[]interface{}{
				1,
				2,
			},
		},
		{`let a = 1;
		  a;`,
			[]code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpPop),
			},
			[]interface{}{
				1,
			},
		},
		{`let a = 1;
		  let b = a;
		  b;`,
			[]code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpPop),
			},
			[]interface{}{
				1,
			},
		},
	}

	runTests(t, tests)
}

func TestArray(t *testing.T) {
	tests := []compileTestCase{
		{`[]`,
			[]code.Instructions{
				code.Make(code.OpArray, 0),
				code.Make(code.OpPop),
			},
			[]interface{}{},
		},
		{`[1,2,3][0]`,
			[]code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
			[]interface{}{
				1, 2, 3, 0,
			},
		},
		{`[1, 2 + 15, false, "hello" + "world"]`,
			[]code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpAdd),
				code.Make(code.OpFalse),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpAdd),
				code.Make(code.OpArray, 4),
				code.Make(code.OpPop),
			},
			[]interface{}{
				1, 2, 15, "hello", "world",
			},
		},
		{`[1, 2 + 15, false, "hello" + "world"][26 + 1]`,
			[]code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpAdd),
				code.Make(code.OpFalse),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpAdd),
				code.Make(code.OpArray, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpConstant, 6),
				code.Make(code.OpAdd),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
			[]interface{}{
				1, 2, 15, "hello", "world", 26, 1,
			},
		},
	}

	runTests(t, tests)
}

func TestHash(t *testing.T) {
	tests := []compileTestCase{
		//{`{}`,
		//	[]code.Instructions{
		//		code.Make(code.OpHash, 0),
		//		code.Make(code.OpPop),
		//	},
		//	[]interface{}{},
		//},
		//{`{1:2,3:4}`,
		//	[]code.Instructions{
		//		code.Make(code.OpConstant, 0),
		//		code.Make(code.OpConstant, 1),
		//		code.Make(code.OpConstant, 2),
		//		code.Make(code.OpConstant, 3),
		//		code.Make(code.OpHash, 2),
		//		code.Make(code.OpPop),
		//	},
		//	[]interface{}{
		//		1, 2, 3, 4,
		//	},
		//},
		{`{1:2,3:4,5:6}[3]`,
			[]code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpHash, 3),
				code.Make(code.OpConstant, 6),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
			[]interface{}{
				1, 2, 3, 4, 5, 6, 3,
			},
		},
	}

	runTests(t, tests)
}

func TestFunctions(t *testing.T) {
	tests := []compileTestCase{
		{
			input: `fn() {return 100 + 200}`,
			expectConstants: []interface{}{
				100,
				200,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
			},
			expectInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `fn() {100 + 200}`,
			expectConstants: []interface{}{
				100,
				200,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
			},
			expectInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `fn() { 100;  200}`,
			expectConstants: []interface{}{
				100,
				200,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpPop),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `fn() { }`,
			expectConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpReturn),
				},
			},
			expectInstructions: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `fn() { 25 }()`,
			expectConstants: []interface{}{
				25,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpCall, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `let noArg = fn() { 25 }
					noArg()`,
			expectConstants: []interface{}{
				25,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpCall, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runTests(t, tests)
}

func TestLetStatement(t *testing.T) {
	tests := []compileTestCase{
		{
			input: `let num = 100
				fn () {num}`,
			expectConstants: []interface{}{
				100,
				object.CompiledFunction{
					Instructions: code.FlattenInstructions([]code.Instructions{
						code.Make(code.OpGetGlobal, 0),
						code.Make(code.OpReturnValue),
					}),
					NumLocals: 0,
				},
			},
			expectInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `fn () { let num = 100; num }`,
			expectConstants: []interface{}{
				100,
				object.CompiledFunction{
					Instructions: code.FlattenInstructions([]code.Instructions{
						code.Make(code.OpConstant, 0),
						code.Make(code.OpSetLocal, 0),
						code.Make(code.OpGetLocal, 0),
						code.Make(code.OpReturnValue),
					}),
					NumLocals: 1,
				},
			},
			expectInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `fn () { let a = 100; let b = 200; a + b }`,
			expectConstants: []interface{}{
				100,
				200,
				object.CompiledFunction{
					Instructions: code.FlattenInstructions([]code.Instructions{
						code.Make(code.OpConstant, 0),
						code.Make(code.OpSetLocal, 0),
						code.Make(code.OpConstant, 1),
						code.Make(code.OpSetLocal, 1),
						code.Make(code.OpGetLocal, 0),
						code.Make(code.OpGetLocal, 1),
						code.Make(code.OpAdd),
						code.Make(code.OpReturnValue),
					}),
					NumLocals:     2,
					NumParameters: 0,
				},
			},
			expectInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `let oneArg = fn (a) { }; oneArg(24)`,
			expectConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpReturn),
				},
				24,
			},
			expectInstructions: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: `let manyArgs = fn (a, b, c) { }; manyArgs(24, 25, 26)`,
			expectConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpReturn),
				},
				24,
				25,
				26,
			},
			expectInstructions: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpCall, 3),
				code.Make(code.OpPop),
			},
		},
		{
			input: `let sum = fn (a, b, c) { let d = 1; a + b + c + d;}; sum(24, 25, 26)`,
			expectConstants: []interface{}{
				1,
				object.CompiledFunction{
					Instructions: code.FlattenInstructions([]code.Instructions{
						code.Make(code.OpConstant, 0),
						code.Make(code.OpSetLocal, 3),
						code.Make(code.OpGetLocal, 0),
						code.Make(code.OpGetLocal, 1),
						code.Make(code.OpAdd),
						code.Make(code.OpGetLocal, 2),
						code.Make(code.OpAdd),
						code.Make(code.OpGetLocal, 3),
						code.Make(code.OpAdd),
						code.Make(code.OpReturnValue),
					}),
					NumLocals:     4,
					NumParameters: 3,
				},
				24,
				25,
				26,
			},
			expectInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpCall, 3),
				code.Make(code.OpPop),
			},
		},
	}

	runTests(t, tests)
}

func TestBuiltin(t *testing.T) {
	tests := []compileTestCase{
		{
			input:           `len([]); push([], 1);`,
			expectConstants: []interface{}{1},
			expectInstructions: []code.Instructions{
				code.Make(code.OpGetBuiltin, 0),
				code.Make(code.OpArray, 0),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
				code.Make(code.OpGetBuiltin, 4),
				code.Make(code.OpArray, 0),
				code.Make(code.OpConstant, 0),
				code.Make(code.OpCall, 2),
				code.Make(code.OpPop),
			},
		},
		{
			input: `fn() { len([]) }`,
			expectConstants: []interface{}{
				object.CompiledFunction{
					Instructions: code.FlattenInstructions([]code.Instructions{
						code.Make(code.OpGetBuiltin, 0),
						code.Make(code.OpArray, 0),
						code.Make(code.OpCall, 1),
						code.Make(code.OpReturnValue),
					}),
					NumLocals:     0,
					NumParameters: 0,
				},
			},
			expectInstructions: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runTests(t, tests)
}

func TestClosures(t *testing.T) {
	tests := []compileTestCase{
		{
			input: `fn(a) { fn(b) { a+b } }`,
			expectConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 0, 1),
					code.Make(code.OpReturnValue),
				},},
			expectInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
              fn(a) {
                  fn(b) {
                    fn(c) { a+b+c }
                  }
              };`,
			expectConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpGetFree, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 0, 2),
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 1, 1),
					code.Make(code.OpReturnValue),
				},},
			expectInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
		{

			input: `
               let global = 55;
               fn() {
                   let a = 66;
                   fn() {
                       let b = 77;
                   
                       fn() {
                          let c = 88;
                          global + a + b + c;
                       }        
                   }
               }`,
			expectConstants: []interface{}{
				55,
				66,
				77,
				88,
				[]code.Instructions{
					code.Make(code.OpConstant, 3),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpGetGlobal, 0),
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpAdd),
					code.Make(code.OpGetFree, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpConstant, 2),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 4, 2),
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpConstant, 1),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 5, 1),
					code.Make(code.OpReturnValue)},
			},
			expectInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpClosure, 6, 0),
				code.Make(code.OpPop),
			},
		},
	}
	runTests(t, tests)
}

func TestRecursiveFunctions(t *testing.T) {
	tests := []compileTestCase{
		{
			input: `
           let wrapper = fn() {
               let countDown = fn(x) {
                   if (x == 0) {
                       return 0;
                   } else {
                       countDown(x - 1);
                   } 
               };
               countDown(1);
           };
           wrapper();
           `,
			expectConstants: []interface{}{
				0,
				0,
				1,
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpConstant, 0),
					code.Make(code.OpEqual),
					code.Make(code.OpJumptNotTruethy, 16),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpReturnValue),
					code.Make(code.OpJump, 25),
					code.Make(code.OpCurrentClosure),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpConstant, 2),
					code.Make(code.OpSubtraction),
					code.Make(code.OpCall, 1),
					code.Make(code.OpReturnValue),
				},
				1,
				[]code.Instructions{
					code.Make(code.OpClosure, 3, 0),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpConstant, 4),
					code.Make(code.OpCall, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectInstructions: []code.Instructions{
				code.Make(code.OpClosure, 5, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpCall, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runTests(t, tests)
}
