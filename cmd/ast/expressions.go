package ast

import (
	"bytes"
	"strings"

	"github.com/ohnishat/peepoScript/cmd/token"
)

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return "\"" + sl.Token.Literal + "\"" }

type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (b *BooleanLiteral) expressionNode()      {}
func (b *BooleanLiteral) TokenLiteral() string { return b.Token.Literal }
func (b *BooleanLiteral) String() string       { return b.Token.Literal }

// OperatorNode is a peepoFriendship/PepegaCredit/mitosis/... marker inside a
// PrefixExpression. The evaluator decides what it binds to.
type OperatorNode struct {
	Token    token.Token
	Operator string
}

func (on *OperatorNode) expressionNode()      {}
func (on *OperatorNode) TokenLiteral() string { return on.Token.Literal }
func (on *OperatorNode) String() string       { return on.Token.Literal }

// ArrayLiteral is a Thinking1 ... Thinking2 list literal. Each element is
// its own expression, so commas and brackets never leak into the enclosing
// hunk stream.
type ArrayLiteral struct {
	Token    token.Token // Thinking1
	Closing  token.Token // Thinking2
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer
	out.WriteString(al.Token.Literal)
	for i, e := range al.Elements {
		if i > 0 {
			out.WriteString(",")
		}
		out.WriteString(" ")
		out.WriteString(e.String())
	}
	out.WriteString(" ")
	out.WriteString(al.Closing.Literal)
	return out.String()
}

// IndexNode trails an operand in a hunk stream and indexes into whatever
// the operand evaluated to: `xs Thinking1 0 Thinking2`. Like OperatorNode,
// it only makes sense at eval time.
type IndexNode struct {
	Token   token.Token // Thinking1
	Closing token.Token // Thinking2
	Index   Expression
}

func (in *IndexNode) expressionNode()      {}
func (in *IndexNode) TokenLiteral() string { return in.Token.Literal }
func (in *IndexNode) String() string {
	return in.Token.Literal + " " + in.Index.String() + " " + in.Closing.Literal
}

// PrefixExpression is a flat operator-first expression: `peepoFriendship x y`,
// `add five ten`. Operands and operators are kept in source order; how they
// group is resolved at eval time, because whether an identifier is a call
// depends on what it is bound to (and its arity).
type PrefixExpression struct {
	Token token.Token
	Hunks []Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	hunks := make([]string, 0, len(pe.Hunks))
	for _, h := range pe.Hunks {
		hunks = append(hunks, h.String())
	}
	out.WriteString(strings.Join(hunks, " "))
	return out.String()
}

type FunctionLiteral struct {
	Token      token.Token // SadgeBusiness
	Parameters []*Identifier
	Body       *Program
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer
	params := make([]string, 0, len(fl.Parameters))
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(fl.Token.Literal)
	out.WriteString(" ")
	out.WriteString(strings.Join(params, " "))
	out.WriteString(" Wokege ")
	out.WriteString(fl.Body.String())
	out.WriteString(" Bedge")
	return out.String()
}

type IfExpression struct {
	Token       token.Token // Hmmge
	Condition   Expression
	Consequence *Program
	Alternative *Program
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out bytes.Buffer
	out.WriteString(ie.Token.Literal)
	out.WriteString(" ")
	out.WriteString(ie.Condition.String())
	out.WriteString(" Wokege ")
	out.WriteString(ie.Consequence.String())
	out.WriteString(" Bedge")
	if ie.Alternative != nil {
		out.WriteString(" peepoShrug Wokege ")
		out.WriteString(ie.Alternative.String())
		out.WriteString(" Bedge")
	}
	return out.String()
}

// ForExpression is a peepoJuice loop. Init and Step are optional PepoG or
// peepoCookie statements; with both left out the loop works as a while over
// Condition. Condition must be present, so there is no accidental infinite
// form.
type ForExpression struct {
	Token     token.Token // peepoJuice
	Init      Expression
	Condition Expression
	Step      Expression
	Body      *Program
}

func (fe *ForExpression) expressionNode()      {}
func (fe *ForExpression) TokenLiteral() string { return fe.Token.Literal }
func (fe *ForExpression) String() string {
	var out bytes.Buffer
	out.WriteString(fe.Token.Literal)
	if fe.Init != nil {
		out.WriteString(" ")
		out.WriteString(fe.Init.String())
	}
	out.WriteString(" ")
	out.WriteString(fe.Condition.String())
	out.WriteString(".")
	if fe.Step != nil {
		out.WriteString(" ")
		out.WriteString(fe.Step.String())
	}
	out.WriteString(" Wokege ")
	out.WriteString(fe.Body.String())
	out.WriteString(" Bedge")
	return out.String()
}
