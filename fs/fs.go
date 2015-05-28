package fs

import "os"

// Exists check if file node exist
func Exists(path string) bool {
	_, err := os.Lstat(path)

	return err == nil
}

// IsDir check if file node is a directory
func IsDir(path string) bool {
	fi, err := os.Lstat(path)

	return err != nil && fi.IsDir()
}

// MkdirAll .
var MkdirAll = os.MkdirAll

// SameFile .
func SameFile(f1, f2 string) bool {
	s1, err := os.Stat(f1)

	if err != nil {
		return false
	}

	s2, err := os.Stat(f2)

	if err != nil {
		return false
	}

	return os.SameFile(s1, s2)
}
