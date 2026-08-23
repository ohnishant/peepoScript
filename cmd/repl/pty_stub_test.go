//go:build !linux

package repl

// Stubs so the integration tests compile on platforms without the
// Linux /dev/ptmx ioctl interface. startREPL skips itself when these
// report unsupported.

import (
	"errors"
	"os"
)

func openPty() (*os.File, *os.File, error) {
	return nil, nil, errors.New("pty integration tests are linux-only")
}

func setPtySize(f *os.File, cols, rows uint16) error {
	return errors.New("pty integration tests are linux-only")
}

func spawnInPty(bin string) (*replSession, error) {
	return nil, errors.New("pty integration tests are linux-only")
}
