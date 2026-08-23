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

func TestParseFor(t *testing.T) {
	input := `peepoJuice PepoG i 0. peepoLessThan i 3. peepoCookie i peepoFriendship i 1. Wokege
		peepoChat i.
	Bedge`

	program, errors := Parse(input)
	if len(errors) > 0 {
		t.Fatalf("unexpected parser errors %v", errors)
	}

	forExpr, ok := program.Expressions[0].(*ast.ForExpression)
	if !ok {
		t.Fatalf("expected *ast.ForExpression, got %T", program.Expressions[0])
	}
	if forExpr.Init == nil || forExpr.Step == nil {
		t.Fatal("expected init and step clauses")
	}
	if forExpr.Condition.String() != "peepoLessThan i 3" {
		t.Errorf("wrong condition: %s", forExpr.Condition.String())
	}
	want := "peepoJuice PepoG i 0. peepoLessThan i 3. peepoCookie i peepoFriendship i 1. Wokege peepoChat i Bedge"
	if got := program.String(); got != want {
		t.Errorf("round-trip mismatch:\nwant %q\ngot  %q", want, got)
	}
}

func TestParseForWhileForm(t *testing.T) {
	input := `PepoG n 3.
	peepoJuice n peepoGreaterThan 0. Wokege
		peepoCookie n PepegaCredit n 1.
	Bedge`

	program, errors := Parse(input)
	if len(errors) > 0 {
		t.Fatalf("unexpected parser errors %v", errors)
	}

	forExpr, ok := program.Expressions[1].(*ast.ForExpression)
	if !ok {
		t.Fatalf("expected *ast.ForExpression, got %T", program.Expressions[1])
	}
	if forExpr.Init != nil || forExpr.Step != nil {
		t.Error("while form should have no init or step")
	}
	if forExpr.TokenLiteral() != "peepoJuice" {
		t.Errorf("expected peepoJuice token, got %s", forExpr.TokenLiteral())
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
		"peepoJuice PepoG i 0 i peepoLessThan 3. Wokege NODDERS. Bedge",
		"peepoJuice peepoLessThan 1 2 Wokege NODDERS. Bedge",
		"peepoJuice peepoLessThan 1 2. peepoChat 1. Wokege NOPERS. Bedge",
	}

	for _, input := range tests {
		_, errors := Parse(input)
		if len(errors) == 0 {
			t.Errorf("input %q: expected parser errors, got none", input)
		}
	}
}
