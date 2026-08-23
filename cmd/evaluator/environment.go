package evaluator

import (
	"fmt"
	"io"
	"os"
)

type Environment struct {
	store map[string]Object
	outer *Environment
}

// NewEnvironment returns an environment whose builtins write to stdout.
func NewEnvironment() *Environment {
	return NewEnvironmentWithOut(os.Stdout)
}

// NewEnvironmentWithOut returns an environment where peepoChat writes to
// out. Embedders that are not a terminal process (e.g. a future WASM
// build running in a browser) pass their own sink here.
func NewEnvironmentWithOut(out io.Writer) *Environment {
	env := &Environment{store: make(map[string]Object)}
	for name, builtin := range newBuiltinSet(out) {
		env.store[name] = builtin
	}
	return env
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := &Environment{store: make(map[string]Object), outer: outer}
	return env
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		return e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}

func notFound(name string) *Error {
	return &Error{Message: fmt.Sprintf("identifier not found: %s", name)}
}
