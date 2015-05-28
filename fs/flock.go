package fs

import (
	"errors"
	"io/ioutil"
	"os"
)

// ErrAlreadyLocked is an error
var ErrAlreadyLocked = errors.New("ErrAlreadyLocked")

// FLocker is a file-based lock
type FLocker struct {
	fh *os.File
}

// NewFLocker creates new FLocker-based lock (unlocked first)
func NewFLocker(path string) (FLocker, error) {
	fh, err := os.Open(path)
	if err != nil {
		return FLocker{}, err
	}
	return FLocker{fh: fh}, nil
}

// FLock lock file
func FLock(name string, f func() error) error {

	if !Exists(name) {
		ioutil.WriteFile(name, []byte("FLocker"), 0644)
	}

	FLocker, err := NewFLocker(name)

	if err != nil {
		return err
	}

	FLocker.Lock()

	defer func() { FLocker.Unlock() }()

	return f()
}
