package tasks

import (
	"os"
	"os/exec"

	"github.com/gsdocker/gserrors"
	"github.com/gsmake/gsmake"
)

// TaskAtom .
func TaskAtom(runner *gsmake.Runner, args ...string) error {

	domain := "task"

	if len(args) != 0 {
		domain = args[0]
	}

	if err := os.Setenv("GOPATH", runner.RootFS().DomainDir(domain)); err != nil {
		return gserrors.Newf(err, "set GOPATH error")
	}

	cmd := exec.Command("atom")

	cmd.Stderr = os.Stderr

	cmd.Stdout = os.Stdout

	cmd.Stdin = os.Stdin

	cmd.Dir = runner.StartDir()

	if err := cmd.Run(); err != nil {
		return gserrors.Newf(err, "start atom error")
	}

	return nil
}
