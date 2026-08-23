package main

// Integration test that runs every script in examples/ against the
// compiled binary and compares stdout to its golden file.
//
// The examples are documentation, but nothing kept them honest: an
// example that stops parsing after a grammar change still sits in the
// repo looking authoritative. Each .peepo now has a sibling .out with
// the expected output, and this test fails when they diverge. Editing
// an example means regenerating its .out in the same commit, so the
// diff shows exactly which documented behavior changed.
//
// Like the repl integration tests, it needs a Go toolchain on PATH to
// build the binary; otherwise it skips. It needs no pty and no Linux,
// so it runs everywhere CI does.

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var (
	buildOnce sync.Once
	binPath   string
	buildErr  error
)

// buildPeepo compiles the main binary once per test run.
func buildPeepo() (string, error) {
	buildOnce.Do(func() {
		dir, err := os.MkdirTemp("", "peepo-examples-*")
		if err != nil {
			buildErr = err
			return
		}
		binPath = filepath.Join(dir, "peepo")
		buildErr = exec.Command("go", "build", "-o", binPath, ".").Run()
	})
	return binPath, buildErr
}

func TestIntegrationExamplesMatchGoldenOutput(t *testing.T) {
	bin, err := buildPeepo()
	if err != nil {
		t.Skipf("cannot build peepo binary: %v", err)
	}

	scripts, err := filepath.Glob("examples/*.peepo")
	if err != nil {
		t.Fatal(err)
	}
	if len(scripts) == 0 {
		t.Fatal("no example scripts found under examples/")
	}

	for _, script := range scripts {
		golden := strings.TrimSuffix(script, ".peepo") + ".out"
		want, err := os.ReadFile(golden)
		if err != nil {
			t.Fatalf("%s has no golden output file %s", script, golden)
		}

		t.Run(filepath.Base(script), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cmd := exec.Command(bin, script)
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			if err := cmd.Run(); err != nil {
				t.Fatalf("exit error: %v\nstderr:\n%s", err, stderr.String())
			}
			if stderr.Len() > 0 {
				t.Errorf("wrote to stderr: %q", stderr.String())
			}
			if stdout.String() != string(want) {
				t.Errorf("output mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
			}
		})
	}
}
