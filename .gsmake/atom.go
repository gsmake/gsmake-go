package plugin

import (
	"os"
	"os/exec"

	"github.com/gsdocker/gsmake"
)

// TaskAtom setup package develop enverioment
func TaskAtom(context *gsmake.Runner, args ...string) error {

	context.I("start atom ...")

	context.I("gopath :%s", os.Getenv("GOPATH"))

	cmd := exec.Command("atom", context.StartDir())

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
