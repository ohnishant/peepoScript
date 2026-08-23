package evaluator

import "fmt"

type Environment struct {
	store map[string]Object
	outer *Environment
}

func NewEnvironment() *Environment {
	env := &Environment{store: make(map[string]Object)}
	for name, builtin := range builtins {
		env.store[name] = builtin
	}
	return env
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
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
