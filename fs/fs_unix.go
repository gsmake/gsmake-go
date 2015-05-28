// +build !windows

package fs

import "os"

// windows special const variable defines
const (
	ExeSuffix = ""
)

// RemoveAll .
var RemoveAll = os.RemoveAll

// Symlink .
var Symlink = os.Symlink
