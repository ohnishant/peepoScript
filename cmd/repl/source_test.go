package repl

import (
	"reflect"
	"testing"
)

func TestStaticSourceSuggestsByPrefix(t *testing.T) {
	src := staticSource{words: []string{"peepoCookie", "peepoJuice", "PepoG", "Hmmge"}}

	tests := []struct {
		prefix string
		want   []string
	}{
		{"peepo", []string{"peepoCookie", "peepoJuice"}},
		{"P", []string{"PepoG"}},
		{"H", []string{"Hmmge"}},
		{"zzz", nil},
		{"", nil},
	}

	for _, tt := range tests {
		got := src.Suggestions(tt.prefix)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Suggestions(%q) = %v, want %v", tt.prefix, got, tt.want)
		}
	}
}

func TestTokenSourceBackedByKeywords(t *testing.T) {
	got := newTokenSource().Suggestions("peepoC")
	want := []string{"peepoCookie"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	if len(newTokenSource().Suggestions("")) != 0 {
		t.Error("empty prefix should suggest nothing")
	}
}
