package lexer

import (
	"testing"

	"github.com/ohnishat/peepoScript/cmd/token"
)

func TestNextToken(t *testing.T) {
	input := "peepoCookie peepoAdd ()peepoLBrace peepoRBrace,; PepegaCredit"

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.ASSIGN, "peepoCookie"},
		{token.PLUS, "peepoAdd"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACE, "peepoLBrace"},
		{token.RBRACE, "peepoRBrace"},
		{token.COMMA, ","},
		{token.SEMICOLON, ";"},
		{token.MINUS, "PepegaCredit"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, tt := range tests {
		var tok token.Token = l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - wrong token type. Expected %q got %q", i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - wrong token literal. Expected %q got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}
