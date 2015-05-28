// +build windows

package fs

import (
	"syscall"

	"github.com/gsdocker/gserrors"
)

var win32LockFile, win32UnlockFile = func() (uintptr, uintptr) {

	handle, err := syscall.LoadLibrary("kernel32.dll")
	if err != nil {
		gserrors.Panicf(err, "load kernel32.dll error")
	}

	addr1, err := syscall.GetProcAddress(handle, "LockFile")
	if err != nil {
		gserrors.Panicf(err, "query kernel32.dll#LockFile error")
	}

	addr2, err := syscall.GetProcAddress(handle, "UnlockFile")
	if err != nil {
		gserrors.Panicf(err, "query kernel32.dll#LockFile error")
	}

	return addr1, addr2
}()

// Lock acquires the lock, blocking
func (lock FLocker) Lock() error {

	if flag, _ := lock.TryLock(); flag {
		return nil
	}

	return gserrors.Newf(ErrAlreadyLocked, "already locked :%s", lock.fh.Name())
}

// TryLock acquires the lock, non-blocking
func (lock FLocker) TryLock() (bool, error) {

	r0, _, _ := syscall.Syscall6(win32LockFile, 5, lock.fh.Fd(), 0, 0, 0, 1, 0)

	if 0 != int(r0) {
		return true, nil
	}

	return false, nil
}

// Unlock releases the lock
func (lock FLocker) Unlock() error {

	syscall.Syscall6(win32UnlockFile, 5, lock.fh.Fd(), 0, 0, 0, 1, 0)

	return nil
}
