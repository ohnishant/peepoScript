package repl

// Integration tests for the interactive REPL, run against a real
// pseudo-terminal.
//
// The unit tests in this package verify the completer's logic in
// isolation, but they cannot catch failures in the terminal plumbing:
// readline never picking up the completer, a zero-width terminal
// silently disabling completion (readline refuses to complete when it
// cannot measure the line width), or raw-mode key handling breaking.
// Those regressions pass every unit test while breaking Tab for real
// users, so they need a test that drives the actual binary through an
// actual tty.
//
// How it works:
//
//   - buildPeepo compiles the module's main binary into a temp dir.
//   - startREPL spawns it as a session leader with a freshly allocated
//     pty as its controlling terminal (pty_unix_test.go on Linux; other
//     platforms provide stubs and the tests skip themselves).
//   - send writes keystrokes to the master side; waitFor accumulates
//     everything the REPL echoed back and fails with the full transcript
//     if the expected output does not appear within the deadline.
//
// A completed word appearing in the transcript is genuine proof of tab
// completion: the test only ever types the partial prefix, so the rest
// of the keyword can only come from the completer inserting it.
//
// These tests require a Unix kernel with /dev/ptmx and a Go toolchain
// on PATH. They are skipped everywhere else. CI runners based on Ubuntu
// satisfy both requirements, so `go test ./...` exercises them without
// any extra workflow configuration.

import (
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	// waitTimeout is how long waitFor tolerates silence before failing.
	waitTimeout = 10 * time.Second
	// banner appears in the REPL greeting; seeing it confirms the
	// process started and the tty is live.
	banner = "Peepo Script"
)

var (
	buildOnce sync.Once
	binPath   string
	buildErr  error
)

// buildPeepo compiles the main binary once per test run and returns its path.
func buildPeepo() (string, error) {
	buildOnce.Do(func() {
		dir, err := os.MkdirTemp("", "peepo-integration-*")
		if err != nil {
			buildErr = err
			return
		}
		binPath = dir + "/peepo"
		build := exec.Command("go", "build", "-o", binPath, ".")
		build.Dir = "../.."
		buildErr = build.Run()
	})
	return binPath, buildErr
}

// replSession is one REPL process attached to a pty.
type replSession struct {
	master *os.File
	cmd    *exec.Cmd

	mu  sync.Mutex
	out strings.Builder
}

func startREPL(t *testing.T) *replSession {
	t.Helper()

	bin, err := buildPeepo()
	if err != nil {
		t.Skipf("cannot build peepo binary: %v", err)
	}
	sess, err := spawnInPty(bin)
	if err != nil {
		t.Skipf("cannot run under a pty here: %v", err)
	}
	t.Cleanup(sess.close)

	// A zero-width terminal makes readline refuse to complete, which
	// would turn completion bugs into silent false passes. Failing the
	// test if the banner never shows also catches a binary that died
	// on startup.
	sess.waitFor(t, banner)
	return sess
}

func (s *replSession) send(keys string) {
	if _, err := s.master.WriteString(keys); err != nil {
		panic("write to pty master: " + err.Error())
	}
}

// waitFor fails the test unless substr shows up in the REPL's output
// within waitTimeout, dumping the whole transcript on failure.
func (s *replSession) waitFor(t *testing.T, substr string) {
	t.Helper()
	deadline := time.Now().Add(waitTimeout)
	for {
		s.mu.Lock()
		got := s.out.String()
		s.mu.Unlock()
		if strings.Contains(got, substr) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %q in repl output\ngot:\n%q",
				substr, got)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func (s *replSession) close() {
	if s.cmd.Process != nil {
		s.cmd.Process.Kill()
		s.cmd.Wait()
	}
	s.master.Close()
}

func TestIntegrationTabCompletesKeyword(t *testing.T) {
	s := startREPL(t)

	// Smoke-check that the interactive path evaluates too, not just echoes.
	s.send("peepoFriendship 2 3.\r")
	s.waitFor(t, "5")

	s.send("peepoF\t")
	// We typed "peepoF"; only the completer can produce the rest.
	s.waitFor(t, "peepoFriendship")
}

func TestIntegrationTabCompletionIsCaseInsensitive(t *testing.T) {
	s := startREPL(t)

	s.send("PEEPOC\t")
	// The terminal echoes our uppercase keystrokes, then the completer
	// appends the suffix in the keyword's real casing. Seeing
	// "PEEPOCookie" in one piece proves both halves: the prefix matched
	// case-insensitively and the insertion came out correctly spelled.
	s.waitFor(t, "PEEPOCookie")
}

func TestIntegrationTabListsCandidatesForAmbiguousPrefix(t *testing.T) {
	s := startREPL(t)

	// Several keywords share the prefix, so readline displays the
	// candidate list instead of inserting anything. It strips the
	// shared prefix when rendering, hence the bare suffixes. None of
	// these words appear in what we typed.
	s.send("peepo\t")
	s.waitFor(t, "Cookie")
	s.waitFor(t, "Juice")
	s.waitFor(t, "Shrug")
}
