// +build !windows

package fs

import "syscall"

// Lock acquires the lock, blocking
func (lock FLocker) Lock() error {
	return syscall.Flock(int(lock.fh.Fd()), syscall.LOCK_EX)
}

// TryLock acquires the lock, non-blocking
func (lock FLocker) TryLock() (bool, error) {
	err := syscall.Flock(int(lock.fh.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	switch err {
	case nil:
		return true, nil
	case syscall.EWOULDBLOCK:
		return false, nil
	}
	return false, err
}

// Unlock releases the lock
func (lock FLocker) Unlock() error {
	return syscall.Flock(int(lock.fh.Fd()), syscall.LOCK_UN)
}
