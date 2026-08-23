package complete

import (
	"sort"
	"strings"

	"github.com/ohnishat/peepoScript/cmd/token"
)

// Source produces completion suggestions for a given word prefix.
// Implementations can pull from anywhere: language tokens, environment
// bindings, builtins, files, and so on.
type Source interface {
	Suggestions(prefix string) []string
}

// StaticSource suggests from a fixed word list, ignoring empty prefixes.
type StaticSource struct {
	Words []string
}

func (s StaticSource) Suggestions(prefix string) []string {
	if prefix == "" {
		return nil
	}
	var matches []string
	for _, word := range s.Words {
		if strings.HasPrefix(word, prefix) {
			matches = append(matches, word)
		}
	}
	sort.Strings(matches)
	return matches
}

// NewTokenSource returns the default source backed by the language keywords.
func NewTokenSource() Source {
	return StaticSource{Words: token.Keywords()}
}
