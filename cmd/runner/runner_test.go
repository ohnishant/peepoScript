package runner

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ohnishat/peepoScript/cmd/evaluator"
)

func TestRunPrintsThroughEnvironmentWriter(t *testing.T) {
	var out bytes.Buffer
	env := evaluator.NewEnvironmentWithOut(&out)

	if _, err := Run(`peepoChat "hello" peepoFriendship 1 2.`, env); err != nil {
		t.Fatalf("Run returned error: %s", err)
	}
	if got := out.String(); got != "hello 3\n" {
		t.Errorf("peepoChat wrote %q, want %q", got, "hello 3\n")
	}
}

func TestRunDoesNotEchoFinalValue(t *testing.T) {
	var out bytes.Buffer
	env := evaluator.NewEnvironmentWithOut(&out)

	result, err := Run(`PepoG x 10. x`, env)
	if err != nil {
		t.Fatalf("Run returned error: %s", err)
	}
	if got := result.Inspect(); got != "10" {
		t.Errorf("result = %q, want 10", got)
	}
	if out.String() != "" {
		t.Errorf("script mode echoed to output: %q", out.String())
	}
}

func TestRunStateCarriesAcrossCalls(t *testing.T) {
	var out bytes.Buffer
	env := evaluator.NewEnvironmentWithOut(&out)

	// Two separate Run calls over one environment: embedders keep a
	// single environment alive across invocations (the browser will do
	// this between button clicks), so bindings must survive.
	if _, err := Run(`PepoG answer 42.`, env); err != nil {
		t.Fatalf("first Run returned error: %s", err)
	}
	if _, err := Run(`peepoChat answer.`, env); err != nil {
		t.Fatalf("second Run returned error: %s", err)
	}
	if got := strings.TrimSpace(out.String()); got != "42" {
		t.Errorf("peepoChat wrote %q, want 42", got)
	}
}

func TestRunParseError(t *testing.T) {
	_, err := Run(`PepoG Wokege`, evaluator.NewEnvironment())
	if err == nil {
		t.Fatal("expected a parse error, got nil")
	}
	var parseErr *ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("error is %T, want *ParseError", err)
	}
	if len(parseErr.Errors) == 0 {
		t.Error("ParseError carries no diagnostics")
	}
}

func TestRunRuntimeError(t *testing.T) {
	_, err := Run(`peepoChat missingIdent.`, evaluator.NewEnvironment())
	if err == nil {
		t.Fatal("expected a runtime error, got nil")
	}
	var rtErr *RuntimeError
	if !errors.As(err, &rtErr) {
		t.Fatalf("error is %T, want *RuntimeError", err)
	}
	if !strings.HasPrefix(rtErr.Error(), "Pepega bro, what the fuck are you doing: ") {
		t.Errorf("runtime error = %q, want the Pepega prefix", rtErr.Error())
	}
}

func TestResolvePath(t *testing.T) {
	dir := t.TempDir()

	exact := filepath.Join(dir, "exact.peepo")
	if err := os.WriteFile(exact, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	noExt := filepath.Join(dir, "noext")
	if err := os.WriteFile(noExt+DefaultExtension, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(dir, "other.txt")
	if err := os.WriteFile(other, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		arg  string
		want string
	}{
		{"explicit extension passes through", exact, exact},
		{"default extension appended", noExt, noExt + DefaultExtension},
		{"existing non-peepo file untouched", other, other},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolvePath(tt.arg)
			if err != nil {
				t.Fatalf("ResolvePath(%q) returned error: %s", tt.arg, err)
			}
			if got != tt.want {
				t.Errorf("ResolvePath(%q) = %q, want %q", tt.arg, got, tt.want)
			}
		})
	}

	t.Run("missing script reports both candidates", func(t *testing.T) {
		arg := filepath.Join(dir, "absent")
		_, err := ResolvePath(arg)
		if err == nil {
			t.Fatal("expected an error for a missing script")
		}
		for _, want := range []string{arg, arg + DefaultExtension} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error %q does not mention %q", err.Error(), want)
			}
		}
	})
}
