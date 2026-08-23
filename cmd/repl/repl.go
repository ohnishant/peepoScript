package repl

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/ohnishat/peepoScript/cmd/evaluator"
	"github.com/ohnishat/peepoScript/cmd/parser"
)

const PROMPT = ">> "
const CONTINUATION_PROMPT = "... "

// isIncomplete reports whether every parser error means "keep reading":
// blocks must close with Bedge on a later line.
func isIncomplete(errors []string) bool {
	if len(errors) == 0 {
		return false
	}
	for _, err := range errors {
		if !strings.Contains(err, "missing Bedge") {
			return false
		}
	}
	return true
}

func StartRepl(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := evaluator.NewEnvironment()
	pending := ""

	for {
		if pending == "" {
			fmt.Fprint(out, PROMPT)
		} else {
			fmt.Fprint(out, CONTINUATION_PROMPT)
		}
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		input := pending + "\n" + scanner.Text()

		program, errors := parser.Parse(input)
		if isIncomplete(errors) {
			pending = input
			continue
		}
		pending = ""

		if len(errors) > 0 {
			for _, err := range errors {
				fmt.Fprintln(out, err)
			}
			continue
		}

		evaluated := evaluator.Eval(program, env)
		if evaluated.Type() != evaluator.NULL_OBJ {
			fmt.Fprintln(out, evaluated.Inspect())
		}
	}
}
