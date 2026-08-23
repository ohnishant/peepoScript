package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"golang.org/x/term"

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

type session struct {
	env     *evaluator.Environment
	pending string
}

func (s *session) evalLine(line string) []string {
	input := s.pending + "\n" + line

	program, errors := parser.Parse(input)
	if isIncomplete(errors) {
		s.pending = input
		return nil
	}
	s.pending = ""

	if len(errors) > 0 {
		return errors
	}

	evaluated := evaluator.Eval(program, s.env)
	if evaluated.Type() != evaluator.NULL_OBJ {
		return []string{evaluated.Inspect()}
	}
	return nil
}

// runeCompleter adapts completion sources to readline's tab handler.
type runeCompleter struct {
	sources []Source
}

func (c *runeCompleter) Do(line []rune, pos int) ([][]rune, int) {
	prefix := string(line[:pos])
	start := len(prefix)
	for i := start - 1; i >= 0; i-- {
		if strings.ContainsRune(" \t\n", rune(prefix[i])) {
			break
		}
		start = i
	}
	word := prefix[start:]

	var suggestions []string
	for _, src := range c.sources {
		suggestions = append(suggestions, src.Suggestions(word)...)
	}
	if len(suggestions) == 0 {
		return nil, 0
	}

	items := make([][]rune, len(suggestions))
	for i, s := range suggestions {
		items[i] = []rune(strings.TrimPrefix(s, word))
	}
	return items, start
}

func defaultSources() []Source {
	return []Source{newTokenSource()}
}

// interactive runs the REPL through readline so Tab completes against
// the configured sources.
func interactive(sources []Source) error {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          PROMPT,
		InterruptPrompt: "^C",
	})
	if err != nil {
		return err
	}
	defer rl.Close()
	rl.Config.AutoComplete = &runeCompleter{sources: sources}

	sess := &session{env: evaluator.NewEnvironment()}
	for {
		if sess.pending != "" {
			rl.SetPrompt(CONTINUATION_PROMPT)
		} else {
			rl.SetPrompt(PROMPT)
		}
		line, err := rl.Readline()
		if err != nil { // io.EOF or ^C
			return nil
		}
		for _, out := range sess.evalLine(line) {
			fmt.Println(out)
		}
	}
}

func StartRepl(in io.Reader, out io.Writer) {
	if f, ok := in.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		if err := interactive(defaultSources()); err == nil {
			return
		}
	}

	scanner := bufio.NewScanner(in)
	sess := &session{env: evaluator.NewEnvironment()}

	for {
		if sess.pending == "" {
			fmt.Fprint(out, PROMPT)
		} else {
			fmt.Fprint(out, CONTINUATION_PROMPT)
		}
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		for _, line := range sess.evalLine(scanner.Text()) {
			fmt.Fprintln(out, line)
		}
	}
}
