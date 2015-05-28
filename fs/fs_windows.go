// +build windows

package fs

import (
	"os"
	"os/exec"

	"github.com/gsdocker/gserrors"
)

// windows special const variable defines
const (
	ExeSuffix = ".exe"
)

// RemoveAll .
func RemoveAll(dir string) error {

	cmd := exec.Command("cmd", "/C", "rd", "/S", "/Q", dir)

	output, err := cmd.Output()

	if err != nil {
		return gserrors.Newf(err, string(output))
	}

	return nil
}

// Symlink fix os.Symlink bug on windows platform
var Symlink = os.Symlink

// func Symlink(src, dst string) error {
//
// 	cmd := exec.Command("cmd", "/u", "/c", "mklink", "/D", dst, src)
//
// 	var buff bytes.Buffer
//
// 	cmd.Stderr = &buff
//
// 	err := cmd.Run()
//
// 	if err != nil {
// 		return gserrors.Newf(err, buff.String())
// 	}
// 	return nil
// }
