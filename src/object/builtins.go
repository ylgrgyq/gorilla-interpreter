package object

import "fmt"

var Builtins = []struct {
	Name     string
	Butiltin *Builtin
}{
	{"len", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError(fmt.Sprintf("wrong number of arguments. expect=%d, got=%d", 1, len(args)))
			}

			switch arg := args[0].(type) {
			case *String:
				return &Integer{Value: int64(len(arg.Value))}
			case *Array:
				return &Integer{Value: int64(len(arg.Elements))}
			default:
				return newError(fmt.Sprintf("argument to `len` not supported, got %s", args[0].Type()))
			}
		}}},

	{"first", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError(fmt.Sprintf("wrong number of arguments for function first. expect=%d, got=%d", 1, len(args)))
			}

			array, ok := args[0].(*Array)
			if !ok {
				return newError(fmt.Sprintf("wrong argument passed to function first. expect Array, got=%q", args[0].Type()))
			}

			length := len(array.Elements)
			if length > 0 {
				return array.Elements[0]
			}
			return NULL
		}}},
	{"last", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError(fmt.Sprintf("wrong number of arguments for function last. expect=%d, got=%d", 1, len(args)))
			}

			array, ok := args[0].(*Array)
			if !ok {
				return newError(fmt.Sprintf("wrong argument passed to function last. expect Array, got=%q", args[0].Type()))
			}

			length := len(array.Elements)
			if length > 0 {
				return array.Elements[len(array.Elements)-1]
			}
			return NULL
		}}},
	{"rest", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError(fmt.Sprintf("wrong number of arguments for function rest. expect=%d, got=%d", 1, len(args)))
			}

			array, ok := args[0].(*Array)
			if !ok {
				return newError(fmt.Sprintf("wrong argument passed to function rest. expect Array, got=%q", args[0].Type()))
			}

			length := len(array.Elements)
			if length > 0 {
				return &Array{Elements: array.Elements[1:]}
			}
			return &Array{}
		}}},
	{"push", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError(fmt.Sprintf("wrong number of arguments for function push. expect=%d, got=%d", 2, len(args)))
			}

			array, ok := args[0].(*Array)
			if !ok {
				return newError(fmt.Sprintf("wrong argument passed to function rest. expect Array, got=%q", args[0].Type()))
			}

			length := len(array.Elements)

			newArray := &Array{Elements: make([]Object, length+1, length+1)}
			copy(newArray.Elements, array.Elements)
			newArray.Elements[length] = args[1]

			return newArray
		}}},
}

func FindBuiltinByName(name string) *Builtin {
	for _,b := range Builtins {
		if b.Name == name {
			return b.Butiltin
		}
	}

	return nil
}

func newError(msg string) *Error {
	return &Error{Msg: msg}
}
