package repl

import (
	"reflect"
	"testing"
)

func TestRuneCompleterCompletesLastWord(t *testing.T) {
	c := &runeCompleter{sources: []Source{
		staticSource{words: []string{"peepoCookie", "peepoJuice"}},
	}}

	tests := []struct {
		line    string
		pos     int
		want    [][]rune
		wantPos int
	}{
		{"peepoC", 6, [][]rune{[]rune("ookie")}, 0},
		{"PEEPOC", 6, [][]rune{[]rune("ookie")}, 0},
		{"peepoCookie peepoJ", 18, [][]rune{[]rune("uice")}, 12},
	}

	for _, tt := range tests {
		got, gotPos := c.Do([]rune(tt.line), tt.pos)
		if !reflect.DeepEqual(got, tt.want) || gotPos != tt.wantPos {
			t.Errorf("Do(%q, %d) = %v, %d; want %v, %d",
				tt.line, tt.pos, got, gotPos, tt.want, tt.wantPos)
		}
	}
}

func TestRuneCompleterNoMatches(t *testing.T) {
	c := &runeCompleter{sources: []Source{
		staticSource{words: []string{"peepoCookie"}},
	}}

	if items, _ := c.Do([]rune("zzz"), 3); items != nil {
		t.Errorf("expected no suggestions for zzz, got %v", items)
	}
}
