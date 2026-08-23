package evaluator

import (
	"testing"

	"github.com/ohnishat/peepoScript/cmd/parser"
)

func testEval(t *testing.T, input string) (*Environment, Object) {
	t.Helper()
	env := NewEnvironment()
	program, errors := parser.Parse(input)
	if len(errors) > 0 {
		t.Fatalf("parser errors: %v", errors)
	}
	return env, Eval(program, env)
}

func TestArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"peepoFriendship 2 3.", 5},
		{"PepegaCredit 10 4.", 6},
		{"mitosis 2 3.", 6},
		{"peepoBye 10 2.", 5},
		{"peepoFriendship mitosis 2 3 peepoBye 8 2.", 10},
		{"peepoFriendship peepoFriendship 1 2 peepoFriendship 3 4.", 10},
	}

	for _, tt := range tests {
		_, evaluated := testEval(t, tt.input)
		result, ok := evaluated.(*Integer)
		if !ok {
			t.Fatalf("input %q: expected Integer, got %T", tt.input, evaluated)
		}
		if result.Value != tt.expected {
			t.Errorf("input %q: expected %d, got %d", tt.input, tt.expected, result.Value)
		}
	}
}

func TestStringConcatAndCompare(t *testing.T) {
	_, evaluated := testEval(t, `peepoFriendship "hi " "there".`)
	str, ok := evaluated.(*String)
	if !ok {
		t.Fatalf("expected String, got %T (%v)", evaluated, evaluated)
	}
	if str.Value != "hi there" {
		t.Errorf("expected %q, got %q", "hi there", str.Value)
	}
}

func TestLetAndAssign(t *testing.T) {
	input := `
	PepoG ten 10.
	peepoCookie ten peepoFriendship ten 5.
	`
	env, _ := testEval(t, input)

	val, ok := env.Get("ten")
	if !ok {
		t.Fatal("ten not bound")
	}
	if val.Inspect() != "15" {
		t.Errorf("expected ten == 15, got %s", val.Inspect())
	}
}

func TestFunctions(t *testing.T) {
	input := `
	PepoG add SadgeBusiness x y Wokege
		peepoFriendship x y.
	Bedge
	PepoG result add 40 2.
	`
	env, _ := testEval(t, input)

	val, ok := env.Get("result")
	if !ok {
		t.Fatal("result not bound")
	}
	if val.Inspect() != "42" {
		t.Errorf("expected result == 42, got %s", val.Inspect())
	}
}

func TestNegativeIntegerLiteral(t *testing.T) {
	input := `
	PepoG minusOne PepegaCredit1.
	PepoG subtracted PepegaCredit 10 5.
	PepoG give SadgeBusiness Wokege PepegaCredit1. Bedge
	give.
	`
	env, evaluated := testEval(t, input)

	tests := []struct {
		name     string
		expected string
	}{
		{"minusOne", "-1"},
		{"subtracted", "5"},
	}

	for _, tt := range tests {
		val, ok := env.Get(tt.name)
		if !ok {
			t.Fatalf("%s not bound", tt.name)
		}
		if val.Inspect() != tt.expected {
			t.Errorf("expected %s == %s, got %s", tt.name, tt.expected, val.Inspect())
		}
	}

	if evaluated.Inspect() != "-1" {
		t.Errorf("expected give call to return -1, got %s", evaluated.Inspect())
	}
}

func TestRecursion(t *testing.T) {
	input := `
	PepoG fact SadgeBusiness n Wokege
		Hmmge Scoots n 0 Wokege
			1.
		Bedge peepoShrug Wokege
			mitosis n fact PepegaCredit n 1.
		Bedge
	Bedge
	fact 5.
	`
	_, evaluated := testEval(t, input)
	if evaluated.Inspect() != "120" {
		t.Errorf("expected 120, got %s", evaluated.Inspect())
	}
}

func TestIfElse(t *testing.T) {
	input := `
	PepoG five 5.
	Hmmge peepoLessThan five 20 Wokege
		NODDERS.
	Bedge peepoShrug Wokege
		NOPERS.
	Bedge
	`
	_, evaluated := testEval(t, input)
	if evaluated != TRUE {
		t.Errorf("expected NODDERS, got %s", evaluated.Inspect())
	}
}

func TestNot(t *testing.T) {
	_, evaluated := testEval(t, "NODDERS peepoJuice.")
	if evaluated != FALSE {
		t.Errorf("expected NOPERS, got %s", evaluated.Inspect())
	}
}

func TestBlockScoping(t *testing.T) {
	input := `
	PepoG x 1.
	Hmmge NODDERS Wokege
		PepoG inner 99.
	Bedge
	x.
	`
	_, evaluated := testEval(t, input)
	if evaluated.Inspect() != "1" {
		t.Errorf("expected outer x to stay 1, got %s", evaluated.Inspect())
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		input   string
		message string
	}{
		{"peepoBye 1 0.", "division by zero"},
		{"peepoFriendship 1 NOPERS.", "matching operands"},
		{"missing.", "identifier not found: missing"},
		{"add 1 2.", "identifier not found: add"},
		{"PepoG five 5. peepoFriendship five.", "ran out of operands"},
	}

	for _, tt := range tests {
		_, evaluated := testEval(t, tt.input)
		err, ok := evaluated.(*Error)
		if !ok {
			t.Fatalf("input %q: expected Error, got %T", tt.input, evaluated)
		}
		if !contains(err.Message, tt.message) {
			t.Errorf("input %q: expected message containing %q, got %q", tt.input, tt.message, err.Message)
		}
	}
}

// TestLexerProgram runs the program from the lexer tests end to end and
// checks the resulting environment.
func TestLexerProgram(t *testing.T) {
	input := `
	PepoG ten 10.
	PepoG five 5.

	PepoG add SadgeBusiness x y Wokege
		peepoFriendship x y.
	Bedge

	peepoCookie five peepoFriendship five 5.

	Hmmge peepoLessThan five 20 Wokege
		PepoG result add five ten.
		result.
	Bedge
	`
	env, last := testEval(t, input)

	if last.Type() == ERROR_OBJ {
		t.Fatalf("program failed with error: %s", last.Inspect())
	}
	if last.Inspect() != "20" {
		t.Errorf("expected block to yield 20 (five is 10 by then), got %v", last.Inspect())
	}
	if five, ok := env.Get("five"); !ok || five.Inspect() != "10" {
		t.Errorf("expected five == 10, got %v", five)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
