package ast

import (
	"testing"

	"github.com/ohnishat/peepoScript/cmd/token"
)

func TestLetExpression(t *testing.T) {
	program := &Program{
		Expressions: []Expression{
			&LetExpression{
				Token:       token.Token{Type: token.LET, Literal: "PepoG"},
				Variable:    &Identifier{Token: token.Token{Type: token.IDENT, Literal: "ten"}, Value: "10"},
				AssignValue: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "10"},
			},
		},
	}

	for i := 0; i < len(program.String()); i++ {
		if []byte(program.String())[i] != []byte("PepoG ten x.")[i] {
			t.Errorf("First characters differ: %v != %v", []byte(program.String())[i], []byte("PepoG ten x.")[i])
		}
	}

}
