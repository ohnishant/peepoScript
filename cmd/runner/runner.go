// Package runner executes peepoScript source without any terminal
// plumbing. It is the embedder-facing entry point: the CLI uses it for
// -c and script files, and a future WASM build targets the browser
// through this same API instead of growing its own eval path.
package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ohnishat/peepoScript/cmd/ast"
	"github.com/ohnishat/peepoScript/cmd/evaluator"
	"github.com/ohnishat/peepoScript/cmd/parser"
)

// DefaultExtension is appended to script paths that have no extension,
// so `peepo greet` finds greet.peepo.
const DefaultExtension = ".peepo"

// ParseError carries the parser's diagnostics for one execution.
type ParseError struct {
	Errors []string
}

func (e *ParseError) Error() string {
	return strings.Join(e.Errors, "\n")
}

// RuntimeError wraps an evaluator Error object so callers can tell it
// apart from parse failures while still getting the PepegaAyyy message.
type RuntimeError struct {
	Object *evaluator.Error
}

func (e *RuntimeError) Error() string {
	return e.Object.Inspect()
}

// Run parses code and evaluates every top-level expression in env.
// Unlike the REPL it does not echo the final value back; scripts say
// what they mean through peepoChat. Output goes wherever env's writer
// points (stdout natively, a JS callback once this runs in a browser).
func Run(code string, env *evaluator.Environment) (evaluator.Object, error) {
	program, errs := parseProgram(code)
	if len(errs) > 0 {
		return evaluator.NULL, &ParseError{Errors: errs}
	}

	result := evaluator.Eval(program, env)
	if errObj, ok := result.(*evaluator.Error); ok {
		return result, &RuntimeError{Object: errObj}
	}
	return result, nil
}

// parseProgram is a thin parser.Parse wrapper, kept as a named seam in
// case the grammar grows multi-file constructs.
func parseProgram(code string) (*ast.Program, []string) {
	return parser.Parse(code)
}

// ResolvePath maps a CLI argument to a script path. Paths that already
// end in the default extension pass through untouched; anything else
// gets the extension appended only when the bare path does not exist,
// so `peepo build.sh` keeps working for non-peepo files.
func ResolvePath(arg string) (string, error) {
	if filepath.Ext(arg) == DefaultExtension {
		return arg, nil
	}
	if _, err := os.Stat(arg); err == nil {
		return arg, nil
	}
	withExt := arg + DefaultExtension
	if _, err := os.Stat(withExt); err == nil {
		return withExt, nil
	}
	return "", fmt.Errorf("no such script: %s (also tried %s)", arg, withExt)
}
