package ast

import "github.com/ohnishat/peepoScript/cmd/token"

// TODO: Figure out what Value actually should be
type Identifier struct {
	Node
	Token token.Token
	Value string
}

func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}

func (i *Identifier) String() string {
	return i.Token.Literal
}

func (i *Identifier) expressionNode() {

}
