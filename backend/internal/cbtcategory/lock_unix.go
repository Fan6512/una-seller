//go:build !windows

package cbtcategory

import (
	"os"
	"syscall"
)

// LockFile takes a non-blocking cross-process lock on a stable sidecar path.
func LockFile(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, err
	}
	// The OS releases this lock when the process or file closes.
	return f, nil
}

func lockUpdate(path string) (*os.File, error) {
	return LockFile(path)
}
