package evaluator

import (
	"fmt"
	"strings"
)

var builtins = map[string]*Builtin{
	"peepoChat": {Arity: -1, Fn: func(args ...Object) Object {
		parts := make([]string, 0, len(args))
		for _, arg := range args {
			parts = append(parts, arg.Inspect())
		}
		fmt.Println(strings.Join(parts, " "))
		return NULL
	}},
	"peepoMeasure": {Arity: 1, Fn: func(args ...Object) Object {
		switch arg := args[0].(type) {
		case *String:
			return &Integer{Value: int64(len(arg.Value))}
		default:
			return &Error{Message: fmt.Sprintf("peepoMeasure needs a string, got %s", arg.Type())}
		}
	}},
}
