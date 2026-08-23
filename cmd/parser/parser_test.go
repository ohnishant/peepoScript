package parser

import (
	"testing"

	"github.com/ohnishat/peepoScript/cmd/ast"
	"github.com/ohnishat/peepoScript/cmd/token"
)

func TestParseExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"five.", "five"},
		{"peepoFriendship five ten.", "peepoFriendship five ten"},
		{"mitosis peepoFriendship 2 3 4.", "mitosis peepoFriendship 2 3 4"},
		{"PepegaCredit ten 5.", "PepegaCredit ten 5"},
		{"peepoBye 10 mitosis 2 2.", "peepoBye 10 mitosis 2 2"},
		{"Scoots five 5.", "Scoots five 5"},
		{"peepoLessThan five 20.", "peepoLessThan five 20"},
		{"NODDERS peepoJuice.", "NODDERS peepoJuice"},
		{"add five ten.", "add five ten"},
		{"add add 1 2 3.", "add add 1 2 3"},
	}

	for _, tt := range tests {
		program, errors := Parse(tt.input)
		if len(errors) > 0 {
			t.Fatalf("input %q: unexpected parser errors %v", tt.input, errors)
		}
		if len(program.Expressions) != 1 {
			t.Fatalf("input %q: expected 1 statement, got %d", tt.input, len(program.Expressions))
		}
		if got := program.String(); got != tt.expected {
			t.Errorf("input %q: expected %q, got %q", tt.input, tt.expected, got)
		}
	}
}

func TestParseLetAndAssign(t *testing.T) {
	program, errors := Parse("PepoG ten 10.")
	if len(errors) > 0 {
		t.Fatalf("unexpected parser errors %v", errors)
	}
	let, ok := program.Expressions[0].(*ast.LetExpression)
	if !ok {
		t.Fatalf("expected *ast.LetExpression, got %T", program.Expressions[0])
	}
	name := let.Variable.(*ast.Identifier)
	if name.Value != "ten" {
		t.Errorf("expected variable name ten, got %s", name.Value)
	}
	if let.AssignValue.String() != "10" {
		t.Errorf("expected value 10, got %s", let.AssignValue.String())
	}
}

func TestParseFunctionLiteral(t *testing.T) {
	input := `PepoG add SadgeBusiness x y Wokege
		peepoFriendship x y.
	Bedge`

	program, errors := Parse(input)
	if len(errors) > 0 {
		t.Fatalf("unexpected parser errors %v", errors)
	}

	let, ok := program.Expressions[0].(*ast.LetExpression)
	if !ok {
		t.Fatalf("expected *ast.LetExpression, got %T", program.Expressions[0])
	}
	fn, ok := let.AssignValue.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("expected *ast.FunctionLiteral, got %T", let.AssignValue)
	}
	if len(fn.Parameters) != 2 || fn.Parameters[0].Value != "x" || fn.Parameters[1].Value != "y" {
		t.Errorf("wrong parameters: %v", fn.Parameters)
	}
	if fn.Token.Type != token.FUNCTION {
		t.Errorf("expected function token, got %s", fn.Token.Type)
	}
	if len(fn.Body.Expressions) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(fn.Body.Expressions))
	}
}

func TestParseIfElse(t *testing.T) {
	input := "Hmmge peepoLessThan five 20 Wokege NODDERS Bedge peepoShrug Wokege NOPERS Bedge"

	program, errors := Parse(input)
	if len(errors) > 0 {
		t.Fatalf("unexpected parser errors %v", errors)
	}

	ifExpr, ok := program.Expressions[0].(*ast.IfExpression)
	if !ok {
		t.Fatalf("expected *ast.IfExpression, got %T", program.Expressions[0])
	}
	if ifExpr.Condition.String() != "peepoLessThan five 20" {
		t.Errorf("wrong condition: %s", ifExpr.Condition.String())
	}
	if ifExpr.Alternative == nil {
		t.Fatal("expected an alternative branch")
	}
}

func TestParserErrors(t *testing.T) {
	tests := []string{
		"PepoG",
		"Wokege NODDERS",
		"@#$%",
	}

	for _, input := range tests {
		_, errors := Parse(input)
		if len(errors) == 0 {
			t.Errorf("input %q: expected parser errors, got none", input)
		}
	}
}
