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

func TestNextTokenUnterminatedString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []token.Token
	}{
		{
			name:  "text then end of input",
			input: `"oops`,
			want: []token.Token{
				{Type: token.ILLEGAL, Literal: "oops"},
				{Type: token.EOF, Literal: ""},
			},
		},
		{
			name:  "only an opening quote",
			input: `"`,
			want: []token.Token{
				{Type: token.ILLEGAL, Literal: ""},
				{Type: token.EOF, Literal: ""},
			},
		},
		{
			name:  "mid-program",
			input: `PepoG x "oops.`,
			want: []token.Token{
				{Type: token.LET, Literal: "PepoG"},
				{Type: token.IDENT, Literal: "x"},
				{Type: token.ILLEGAL, Literal: "oops."},
				{Type: token.EOF, Literal: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			for i, want := range tt.want {
				tok := l.NextToken()

				if tok.Type != want.Type {
					t.Fatalf("tests[%d] - wrong token type. Expected %q got %q", i, want.Type, tok.Type)
				}

				if tok.Literal != want.Literal {
					t.Fatalf("tests[%d] - wrong token literal. Expected %q got %q", i, want.Literal, tok.Literal)
				}
			}
		})
	}
}

// A regression here used to spin forever instead of returning, so pull a few
// tokens past the bad input and make sure every call comes back as EOF.
func TestNextTokenKeepsReturningEOFAfterUnterminatedString(t *testing.T) {
	l := New(`"oops`)
	l.NextToken() // ILLEGAL

	for i := 0; i < 5; i++ {
		tok := l.NextToken()
		if tok.Type != token.EOF {
			t.Fatalf("call %d - expected EOF got %q (%q)", i, tok.Type, tok.Literal)
		}
	}
}

func TestNextTokenComplex_1(t *testing.T) {
	input := `
	PepoG ten 10.
	PepoG five 5.

	PepoG add SadgeBusiness x y Wokege
		peepoFriendship x y.
	Bedge
	
	Hmmge NODDERS Wokege
		PepoG result add five ten.
	Bedge

	peepoCookie five peepoFriendship 5.

	Hmmge peepoLessThan five 20 Wokege
		NODDERS		
	Bedge peepoShrug Wokege.
		NOPERS
	Bedge
	"hi i am a string"
	`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LET, "PepoG"},              // 0
		{token.IDENT, "ten"},              // 1
		{token.INT, "10"},                 // 2
		{token.FULLSTOP, "."},             // 3
		{token.LET, "PepoG"},              // 4
		{token.IDENT, "five"},             // 5
		{token.INT, "5"},                  // 6
		{token.FULLSTOP, "."},             // 7
		{token.LET, "PepoG"},              // 8
		{token.IDENT, "add"},              // 9
		{token.FUNCTION, "SadgeBusiness"}, // 10
		{token.IDENT, "x"},                // 11
		{token.IDENT, "y"},                // 12
		{token.LBRACE, "Wokege"},          // 13
		{token.PLUS, "peepoFriendship"},   // 14
		{token.IDENT, "x"},                // 15
		{token.IDENT, "y"},                // 16
		{token.FULLSTOP, "."},             // 17
		{token.RBRACE, "Bedge"},           // 18
		{token.IF, "Hmmge"},               // 19
		{token.TRUE, "NODDERS"},           // 20
		{token.LBRACE, "Wokege"},          // 21
		{token.LET, "PepoG"},              // 22
		{token.IDENT, "result"},           // 23
		{token.IDENT, "add"},              // 24
		{token.IDENT, "five"},             // 25
		{token.IDENT, "ten"},              // 26
		{token.FULLSTOP, "."},             // 27
		{token.RBRACE, "Bedge"},           // 28
		//peepoCookie five peepoFriendship 5.
		{token.ASSIGN, "peepoCookie"},
		{token.IDENT, "five"},
		{token.PLUS, "peepoFriendship"},
		{token.INT, "5"},
		{token.FULLSTOP, "."},
		//
		{token.IF, "Hmmge"},               // 29
		{token.LESSTHAN, "peepoLessThan"}, // 30
		{token.IDENT, "five"},             // 31
		{token.INT, "20"},                 // 32
		{token.LBRACE, "Wokege"},          // 33
		{token.TRUE, "NODDERS"},           // 34
		{token.RBRACE, "Bedge"},           // 35
		{token.ELSE, "peepoShrug"},        // 36
		{token.LBRACE, "Wokege"},          // 37
		{token.FULLSTOP, "."},             // 38
		{token.FALSE, "NOPERS"},           // 39
		{token.RBRACE, "Bedge"},           // 40
		{token.STRING, "hi i am a string"},
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
