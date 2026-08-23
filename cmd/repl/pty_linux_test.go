//go:build linux

package repl

// Linux pty plumbing for the integration tests in
// repl_integration_test.go. Stdlib only: /dev/ptmx plus three ioctls,
// no external pty package.

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

// openPty allocates a fresh pseudo-terminal pair.
func openPty() (*os.File, *os.File, error) {
	master, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		return nil, nil, err
	}

	var n uint32
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, master.Fd(),
		uintptr(syscall.TIOCGPTN), uintptr(unsafe.Pointer(&n))); errno != 0 {
		master.Close()
		return nil, nil, errno
	}

	unlock := int32(0)
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, master.Fd(),
		uintptr(syscall.TIOCSPTLCK), uintptr(unsafe.Pointer(&unlock))); errno != 0 {
		master.Close()
		return nil, nil, errno
	}

	slave, err := os.OpenFile(fmt.Sprintf("/dev/pts/%d", n),
		os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		master.Close()
		return nil, nil, err
	}
	return master, slave, nil
}

// setPtySize gives the terminal a real window size. readline refuses to
// complete when the width is zero, so this is load-bearing for the tests.
func setPtySize(f *os.File, cols, rows uint16) error {
	ws := struct{ rows, cols, x, y uint16 }{rows, cols, 0, 0}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(),
		uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(&ws)))
	if errno != 0 {
		return errno
	}
	return nil
}

// spawnInPty runs bin as a session leader with the pty slave as its
// controlling terminal, so readline enters raw mode exactly like it
// would in a user's terminal. It also starts pumping output from the
// master side into the session buffer.
func spawnInPty(bin string) (*replSession, error) {
	master, slave, err := openPty()
	if err != nil {
		return nil, err
	}
	if err := setPtySize(master, 80, 24); err != nil {
		master.Close()
		slave.Close()
		return nil, err
	}

	cmd := exec.Command(bin)
	cmd.Stdin = slave
	cmd.Stdout = slave
	cmd.Stderr = slave
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,
		Setctty: true,
	}
	if err := cmd.Start(); err != nil {
		master.Close()
		slave.Close()
		return nil, err
	}
	// The child holds its own descriptor; we only need ours.
	slave.Close()

	s := &replSession{master: master, cmd: cmd}
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := master.Read(buf)
			if n > 0 {
				s.mu.Lock()
				s.out.Write(buf[:n])
				s.mu.Unlock()
			}
			if err != nil {
				return
			}
		}
	}()
	return s, nil
}
