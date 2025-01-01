package ast

import (
	"bytes"

	"github.com/ohnishat/peepoScript/cmd/token"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Expressions []Expression
}

func (p *Program) TokenLiteral() string {
	if len(p.Expressions) > 0 {
		return p.Expressions[0].TokenLiteral()
	} else {
		return ""
	}
}

func (p *Program) String() string {
	var str bytes.Buffer

	for _, j := range p.Expressions {
		str.WriteString(j.String())
	}

	return str.String()
}

type LetExpression struct {
	Token       token.Token
	Variable    Node
	AssignValue Expression
}

func (e *LetExpression) expressionNode() {}

func (e *LetExpression) TokenLiteral() string {
	return e.Token.Literal
}

func (e *LetExpression) String() string {
	var str bytes.Buffer
	str.WriteString(e.Token.Literal + " ")
	str.WriteString(e.Variable.String() + " ")
	str.WriteString(e.AssignValue.String())
	str.WriteString(".")

	return str.String()
}
