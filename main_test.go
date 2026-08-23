package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCLIModes(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "hello.peepo")
	if err := os.WriteFile(script, []byte("peepoChat \"from file\"."), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		args      []string
		wantOut   string
		wantErr   string
		wantExit  int
		skipStdin bool
	}{
		{
			name:     "inline code",
			args:     []string{"-c", `peepoChat peepoFriendship 1 2.`},
			wantOut:  "3\n",
			wantExit: 0,
		},
		{
			name:     "script file with extension",
			args:     []string{script},
			wantOut:  "from file\n",
			wantExit: 0,
		},
		{
			name:     "script path without extension resolves to .peepo",
			args:     []string{strings.TrimSuffix(script, ".peepo")},
			wantOut:  "from file\n",
			wantExit: 0,
		},
		{
			name:     "runtime error exits 1 and reports on stderr",
			args:     []string{"-c", "peepoChat nope."},
			wantErr:  "Pepega bro, what the fuck are you doing: identifier not found: nope\n",
			wantExit: 1,
		},
		{
			name:     "parse error exits 1 and reports on stderr",
			args:     []string{"-c", "PepoG Wokege"},
			wantErr:  "Sadge...",
			wantExit: 1,
		},
		{
			name:     "missing script exits 1",
			args:     []string{filepath.Join(dir, "absent")},
			wantErr:  "no such script",
			wantExit: 1,
		},
		{
			name:     "too many args exits 2",
			args:     []string{"a.peepo", "b.peepo"},
			wantErr:  "at most one script file",
			wantExit: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			got := run(tt.args, strings.NewReader(""), &out, &errOut)
			if got != tt.wantExit {
				t.Errorf("exit = %d, want %d (stderr: %q)", got, tt.wantExit, errOut.String())
			}
			if tt.wantOut != "" && out.String() != tt.wantOut {
				t.Errorf("stdout = %q, want %q", out.String(), tt.wantOut)
			}
			if tt.wantErr != "" && !strings.Contains(errOut.String(), tt.wantErr) {
				t.Errorf("stderr = %q, want it to contain %q", errOut.String(), tt.wantErr)
			}
		})
	}
}

func TestRunNoArgsPrintsBanner(t *testing.T) {
	var out, errOut bytes.Buffer
	got := run(nil, strings.NewReader(""), &out, &errOut)
	if got != 0 {
		t.Errorf("exit = %d, want 0", got)
	}
	if !strings.Contains(out.String(), "REPL") {
		t.Errorf("banner missing from stdout: %q", out.String())
	}
}
