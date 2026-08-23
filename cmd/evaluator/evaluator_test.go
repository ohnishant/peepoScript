package evaluator

import (
	"io"
	"testing"

	"github.com/ohnishat/peepoScript/cmd/parser"
)

func testEval(t *testing.T, input string) (*Environment, Object) {
	t.Helper()
	env := NewEnvironmentWithOut(io.Discard)
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
	PepoG minusOne -1.
	PepoG subtracted PepegaCredit 10 5.
	PepoG negatedSum - peepoFriendship 2 3.
	PepoG give SadgeBusiness Wokege -1. Bedge
	give.
	`
	env, evaluated := testEval(t, input)

	tests := []struct {
		name     string
		expected string
	}{
		{"minusOne", "-1"},
		{"subtracted", "5"},
		{"negatedSum", "-5"},
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

func TestForLoop(t *testing.T) {
	input := `
	PepoG total 0.
	peepoJuice PepoG i 0. peepoLessThan i 5. peepoCookie i peepoFriendship i 1. Wokege
		peepoCookie total peepoFriendship total i.
	Bedge
	total.
	`
	_, evaluated := testEval(t, input)
	if evaluated.Inspect() != "10" {
		t.Errorf("expected total == 10, got %s", evaluated.Inspect())
	}
}

func TestForLoopWhileForm(t *testing.T) {
	input := `
	PepoG n 3.
	peepoJuice peepoGreaterThan n 0. peepoCookie n PepegaCredit n 1. Wokege
		peepoChat "tick".
	Bedge
	n.
	`
	_, evaluated := testEval(t, input)

	if evaluated.Inspect() != "0" {
		t.Errorf("expected countdown to leave n at 0, got %s", evaluated.Inspect())
	}
}

func TestForLoopCountsDownWithStep(t *testing.T) {
	input := `
	PepoG last -1.
	peepoJuice PepoG n 3. peepoGreaterThan n 0. peepoCookie n PepegaCredit n 1. Wokege
		peepoChat n.
	Bedge
	`
	_, evaluated := testEval(t, input)
	if evaluated.Type() == ERROR_OBJ {
		t.Fatalf("loop errored: %s", evaluated.Inspect())
	}
}

func TestForLoopScoping(t *testing.T) {
	input := `
	peepoJuice PepoG i 0. peepoLessThan i 2. peepoCookie i peepoFriendship i 1. Wokege
		NODDERS.
	Bedge
	i.
	`
	_, evaluated := testEval(t, input)
	err, ok := evaluated.(*Error)
	if !ok || !contains(err.Message, "identifier not found: i") {
		t.Errorf("expected loop variable to stay scoped, got %v", evaluated.Inspect())
	}
}

func TestListLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Thinking1 Thinking2.", "[]"},
		{"Thinking1 1, 2, 3 Thinking2.", "[1, 2, 3]"},
		{"Thinking1 \"a\", NODDERS, 2 Thinking2.", "[a, NODDERS, 2]"},
		{"Thinking1 1, Thinking1 2, 3 Thinking2 Thinking2.", "[1, [2, 3]]"},
	}

	for _, tt := range tests {
		_, evaluated := testEval(t, tt.input)
		if evaluated.Type() == ERROR_OBJ {
			t.Fatalf("input %q: %s", tt.input, evaluated.Inspect())
		}
		if evaluated.Inspect() != tt.expected {
			t.Errorf("input %q: expected %s, got %s", tt.input, tt.expected, evaluated.Inspect())
		}
	}
}

func TestListIndexing(t *testing.T) {
	input := `
	PepoG xs Thinking1 10, 20, 30 Thinking2.
	xs Thinking1 0 Thinking2.
	`
	_, evaluated := testEval(t, input)
	if evaluated.Inspect() != "10" {
		t.Errorf("expected first element 10, got %s", evaluated.Inspect())
	}

	chained := `
	PepoG nested Thinking1 1, Thinking1 5, 6 Thinking2 Thinking2.
	nested Thinking1 1 Thinking2 Thinking1 0 Thinking2.
	`
	_, evaluated = testEval(t, chained)
	if evaluated.Inspect() != "5" {
		t.Errorf("expected chained index to yield 5, got %s", evaluated.Inspect())
	}

	inExpr := `
	PepoG xs Thinking1 7, 8 Thinking2.
	peepoFriendship xs Thinking1 1 Thinking2 1.
	`
	_, evaluated = testEval(t, inExpr)
	if evaluated.Inspect() != "9" {
		t.Errorf("expected indexed operand in expression to yield 9, got %s", evaluated.Inspect())
	}
}

func TestListIndexErrors(t *testing.T) {
	tests := []struct {
		input   string
		message string
	}{
		{"PepoG xs Thinking1 1 Thinking2. xs Thinking1 5 Thinking2.", "index out of range"},
		{"PepoG xs Thinking1 1 Thinking2. xs Thinking1 \"zero\" Thinking2.", "list index must be an integer"},
		{"PepoG x 1. x Thinking1 0 Thinking2.", "indexing needs a list"},
		{"Thinking1 1 Thinking2 Thinking1 3 Thinking2.", "index out of range"},
	}

	for _, tt := range tests {
		_, evaluated := testEval(t, tt.input)
		err, ok := evaluated.(*Error)
		if !ok {
			t.Fatalf("input %q: expected Error, got %T (%v)", tt.input, evaluated, evaluated)
		}
		if !contains(err.Message, tt.message) {
			t.Errorf("input %q: expected message containing %q, got %q", tt.input, tt.message, err.Message)
		}
	}
}

func TestLoopOverList(t *testing.T) {
	input := `
	PepoG xs Thinking1 3, 1, 4, 1, 5 Thinking2.
	PepoG total 0.
	peepoJuice PepoG i 0. peepoLessThan i peepoMeasure xs. peepoCookie i peepoFriendship i 1. Wokege
		peepoCookie total peepoFriendship total xs Thinking1 i Thinking2.
	Bedge
	total.
	`
	_, evaluated := testEval(t, input)
	if evaluated.Inspect() != "14" {
		t.Errorf("expected sum of list elements 14, got %s", evaluated.Inspect())
	}
}

func TestLoopBodyErrorsStopTheLoop(t *testing.T) {
	input := `
	peepoJuice PepoG i 0. peepoLessThan i 10. peepoCookie i peepoFriendship i 1. Wokege
		peepoBye 1 0.
	Bedge
	`
	_, evaluated := testEval(t, input)
	err, ok := evaluated.(*Error)
	if !ok {
		t.Fatalf("expected body error to propagate, got %v", evaluated.Inspect())
	}
	if !contains(err.Message, "division by zero") {
		t.Errorf("expected division by zero, got %q", err.Message)
	}
}

// TestListEqualityNotStructural pins current behavior tracked in #28:
// equals() has no Array case, so even identical lists compare NOPERS.
// When structural equality lands, flip this to expect TRUE.
func TestListEqualityNotStructural(t *testing.T) {
	input := `
	PepoG xs Thinking1 1, 2 Thinking2.
	PepoG ys Thinking1 1, 2 Thinking2.
	Scoots xs ys.
	`
	_, evaluated := testEval(t, input)
	if evaluated != FALSE {
		t.Errorf("expected identical lists to compare NOPERS (issue #28), got %s", evaluated.Inspect())
	}
}

// TestIndexingCallResults pins how index hunks interact with calls,
// tracked in #29. Zero-arity calls return a value the index applies to;
// higher-arity calls try to consume the raw IndexNode hunk as an
// argument, which currently surfaces as an "unknown expression" error.
// Both halves need revisiting when #29 is decided.
func TestIndexingCallResults(t *testing.T) {
	zeroArity := `
	PepoG make SadgeBusiness Wokege Thinking1 7, 8 Thinking2 Bedge
	make Thinking1 0 Thinking2.
	`
	_, evaluated := testEval(t, zeroArity)
	if evaluated.Inspect() != "7" {
		t.Errorf("expected zero-arity call result to be indexable at 7, got %s", evaluated.Inspect())
	}

	higherArity := `
	PepoG first SadgeBusiness x Wokege x Bedge
	first Thinking1 0 Thinking2 5.
	`
	_, evaluated = testEval(t, higherArity)
	err, ok := evaluated.(*Error)
	if !ok {
		t.Fatalf("expected arity>=1 call followed by an index to error, got %v", evaluated.Inspect())
	}
	if !contains(err.Message, "unknown expression") {
		t.Errorf("expected unknown-expression error (issue #29), got %q", err.Message)
	}
}

// TestStringIndexingUnsupported pins current behavior tracked in #30:
// applyIndex rejects non-list targets rather than returning characters.
func TestStringIndexingUnsupported(t *testing.T) {
	_, evaluated := testEval(t, `"hello" Thinking1 0 Thinking2.`)
	err, ok := evaluated.(*Error)
	if !ok {
		t.Fatalf("expected string indexing to error, got %v", evaluated.Inspect())
	}
	if !contains(err.Message, "indexing needs a list") {
		t.Errorf("expected indexing-needs-a-list error (issue #30), got %q", err.Message)
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
		{"peepoJuice 1. Wokege NOPERS. Bedge", "peepoJuice condition must be NODDERS/NOPERS"},
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
