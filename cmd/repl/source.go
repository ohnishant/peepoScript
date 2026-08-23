package repl

import (
	"sort"
	"strings"

	"github.com/ohnishat/peepoScript/cmd/evaluator"
	"github.com/ohnishat/peepoScript/cmd/token"
)

// Source produces completion suggestions for a given word prefix.
// The REPL stacks sources; new ones can come from anywhere later,
// e.g. environment bindings or builtins.
type Source interface {
	Suggestions(prefix string) []string
}

// staticSource suggests from a fixed word list, matching prefixes
// case-insensitively. Matches keep the word's original casing so the
// inserted completion is always spelled correctly.
type staticSource struct {
	words []string
}

func (s staticSource) Suggestions(prefix string) []string {
	if prefix == "" {
		return nil
	}
	var matches []string
	for _, word := range s.words {
		if len(word) >= len(prefix) && strings.EqualFold(word[:len(prefix)], prefix) {
			matches = append(matches, word)
		}
	}
	sort.Strings(matches)
	return matches
}

// newTokenSource returns the default source backed by the language
// keywords and builtin function names.
func newTokenSource() Source {
	return staticSource{words: append(token.Keywords(), evaluator.Builtins()...)}
}
