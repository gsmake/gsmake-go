// +build !windows

package nfs

import (
	"os"
	"syscall"
)

// windows special const variable defines
const (
	ExeSuffix = ""
)

// RemoveAll .
var RemoveAll = os.RemoveAll

// Symlink .
var Symlink = os.Symlink

// FLocker file lock
func FLocker(file *os.File) error {
	return syscall.FLocker(int(file.Fd()), syscall.LOCK_EX)
}
