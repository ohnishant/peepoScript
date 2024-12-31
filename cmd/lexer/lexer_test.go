package lexer

import (
	"testing"

	"github.com/ohnishat/peepoScript/cmd/token"
)

func TestNextToken_1(t *testing.T) {
	input := "peepoCookie peepoFriendship ()Wokege Bedge,. PepegaCredit"

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.ASSIGN, "peepoCookie"},
		{token.PLUS, "peepoFriendship"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACE, "Wokege"},
		{token.RBRACE, "Bedge"},
		{token.COMMA, ","},
		{token.FULLSTOP, "."},
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
func TestNextTokenComplex_1(t *testing.T) {
	input := `
	PepoG ten 10.
	PepoG five 5.

	PepoG add SadgeBusiness x y Wokege
		peepoFriendship x y.
	Bedge.
	
	Hmmge NODDERS Wokege
		PepoG result add five ten.
	Bedge
	`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LET, "PepoG"},
		{token.IDENT, "ten"},
		{token.INT, "10"},
		{token.FULLSTOP, "."},
		//
		{token.LET, "PepoG"},
		{token.IDENT, "five"},
		{token.INT, "5"},
		{token.FULLSTOP, "."},
		//
		{token.LET, "PepoG"},
		{token.IDENT, "add"},
		{token.FUNCTION, "SadgeBusiness"},
		{token.IDENT, "x"},
		{token.IDENT, "y"},
		{token.LBRACE, "Wokege"},
		//
		{token.PLUS, "peepoFriendship"},
		{token.IDENT, "x"},
		{token.IDENT, "y"},
		{token.FULLSTOP, "."},
		//
		{token.RBRACE, "Bedge"},
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
