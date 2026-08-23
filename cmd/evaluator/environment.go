package evaluator

import (
	"fmt"
	"io"
	"os"
	"sort"
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

// Builtins returns the builtin function names in a stable order.
// Hosts use it for completion and emote lookup alongside Keywords.
func Builtins() []string {
	set := newBuiltinSet(io.Discard)
	names := make([]string, 0, len(set))
	for name := range set {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		return e.outer.Get(name)
	}
	return obj, ok
}

// Set binds name to val. When name already exists in an enclosing scope,
// the binding closest to where the assignment happens is updated, so loop
// bodies and blocks can mutate variables they can read. Genuinely new
// names bind locally.
func (e *Environment) Set(name string, val Object) Object {
	if !e.rebind(name, val) {
		e.store[name] = val
	}
	return val
}

func (e *Environment) rebind(name string, val Object) bool {
	if _, ok := e.store[name]; ok {
		e.store[name] = val
		return true
	}
	if e.outer != nil {
		return e.outer.rebind(name, val)
	}
	return false
}

func notFound(name string) *Error {
	return &Error{Message: fmt.Sprintf("identifier not found: %s", name)}
}
