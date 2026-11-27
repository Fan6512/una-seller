package cbtcategory

import (
	"os"
	"syscall"
)

// LockFile takes a non-blocking cross-process lock on a stable sidecar path.
func LockFile(path string) (*os.File, error) {
	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	handle, err := syscall.CreateFile(name, syscall.GENERIC_READ|syscall.GENERIC_WRITE, 0, nil, syscall.OPEN_ALWAYS, syscall.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return nil, err
	}
	// Opening without a share mode excludes other writers until the handle closes.
	return os.NewFile(uintptr(handle), path), nil
}

func lockUpdate(path string) (*os.File, error) {
	return LockFile(path)
}
