package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENT = "IDENT"
	INT   = "INT"

	ASSIGN   = "ASSIGN"
	PLUS     = "ADD"
	MINUS    = "SUBTRACT"
	MULTIPLY = "MULTIPLY"
	DIVIDE   = "DIVIDE"

	COMMA     = "COMMA"
	SEMICOLON = "SEMICOLON"

	LPAREN = "LPAREN"
	RPAREN = "RPAREN"
	LBRACE = "LBRACE"
	RBRACE = "RBRACE"

	FUNCTION = "FUNCTION"
	LET      = "LET"
	IF       = "IF"
	ELSE     = "ELSE"
)

var keywords = map[string]TokenType{
	"peepoCookie": ASSIGN,

	"peepoLBrace": LBRACE,
	"peepoRBrace": RBRACE,

	"peepoAdd":     PLUS,
	"PepegaCredit": MINUS,
	"mitosis":      MULTIPLY,
	"peepoDivide":  DIVIDE,

	"SadgeBusiness": FUNCTION,
	"PepoG":         LET,

	"Hmmge":      IF,
	"peepoShrug": ELSE,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
