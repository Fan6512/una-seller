//go:build !windows

package cbtcategory

import (
	"os"
	"syscall"
)

func lockUpdate(path string) (*os.File, error) {
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
