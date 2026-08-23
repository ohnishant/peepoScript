package repl

import (
	"sort"
	"strings"

	"github.com/ohnishat/peepoScript/cmd/token"
)

// Source produces completion suggestions for a given word prefix.
// The REPL stacks sources; new ones can come from anywhere later,
// e.g. environment bindings or builtins.
type Source interface {
	Suggestions(prefix string) []string
}

// staticSource suggests from a fixed word list, ignoring empty prefixes.
type staticSource struct {
	words []string
}

func (s staticSource) Suggestions(prefix string) []string {
	if prefix == "" {
		return nil
	}
	var matches []string
	for _, word := range s.words {
		if strings.HasPrefix(word, prefix) {
			matches = append(matches, word)
		}
	}
	sort.Strings(matches)
	return matches
}

// newTokenSource returns the default source backed by the language keywords.
func newTokenSource() Source {
	return staticSource{words: token.Keywords()}
}
