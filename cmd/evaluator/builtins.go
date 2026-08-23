package evaluator

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// builtinSet holds the builtins for one execution. Output goes through
// out instead of a hardcoded stdout, so embedders (REPL today, a WASM
// sandbox on a webpage later) can capture peepoChat wherever they want.
type builtinSet struct {
	out io.Writer
}

func newBuiltinSet(out io.Writer) map[string]*Builtin {
	b := builtinSet{out: out}
	return map[string]*Builtin{
		"peepoChat": {Arity: -1, Fn: b.peepoChat},
		"peepoMeasure": {Arity: 1, Fn: func(args ...Object) Object {
			switch arg := args[0].(type) {
			case *String:
				return &Integer{Value: int64(len(arg.Value))}
			case *Array:
				return &Integer{Value: int64(len(arg.Elements))}
			default:
				return &Error{Message: fmt.Sprintf("peepoMeasure needs a string or list, got %s", arg.Type())}
			}
		}},
	}
}

func (b builtinSet) peepoChat(args ...Object) Object {
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		parts = append(parts, arg.Inspect())
	}
	fmt.Fprintln(b.out, strings.Join(parts, " "))
	return NULL
}

var defaultBuiltins = newBuiltinSet(os.Stdout)
