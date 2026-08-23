package token

import "sort"

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENT = "IDENT"

	INT    = "INT"
	STRING = "STRING"

	ASSIGN   = "ASSIGN"
	PLUS     = "ADD"
	MINUS    = "SUBTRACT"
	NEGATE   = "NEGATE"
	MULTIPLY = "MULTIPLY"
	DIVIDE   = "DIVIDE"

	EQUAL       = "EQUAL"
	LESSTHAN    = "LESSTHAN"
	GREATERTHAN = "GREATERTHAN"
	NOT         = "NOT"

	COMMA    = "COMMA"
	FULLSTOP = "FULLSTOP"

	LPAREN = "LPAREN"
	RPAREN = "RPAREN"
	LBRACE = "LBRACE"
	RBRACE = "RBRACE"

	TRUE  = "TRUE"
	FALSE = "FALSE"

	FUNCTION = "FUNCTION"
	LET      = "LET"
	IF       = "IF"
	ELSE     = "ELSE"
)

var keywords = map[string]TokenType{
	"peepoCookie": ASSIGN,

	"Wokege": LBRACE,
	"Bedge":  RBRACE,

	"peepoFriendship": PLUS,
	"PepegaCredit":    MINUS,
	"mitosis":         MULTIPLY,
	"peepoBye":        DIVIDE, // this is so stupid

	"peepoLessThan":    LESSTHAN,
	"peepoGreaterThan": GREATERTHAN,

	"Scoots":     EQUAL,
	"peepoJuice": NOT,

	"SadgeBusiness": FUNCTION,
	"PepoG":         LET,

	"NODDERS": TRUE,
	"NOPERS":  FALSE,

	"Hmmge":      IF,
	"peepoShrug": ELSE,
}

// Keywords returns the language keyword literals in a stable order.
func Keywords() []string {
	words := make([]string, 0, len(keywords))
	for word := range keywords {
		words = append(words, word)
	}
	sort.Strings(words)
	return words
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
