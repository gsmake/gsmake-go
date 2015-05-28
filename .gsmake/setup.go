package tasks

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gsos"
	"github.com/gsmake/gsmake"
)

// TaskSetup .
func TaskSetup(runner *gsmake.Runner, args ...string) error {

	if runner.Current() == "gsmake://github.com/gsmake/gsmake?domain=task" {

		if len(args) == 0 {
			return gserrors.Newf(nil, "expect setup dir")
		}

		os.Setenv("GOPATH", runner.RootFS().DomainDir("task"))

		obj := filepath.Join(args[0], "bin", "gsmake"+gsos.ExeSuffix)

		runner.I("install gsmake to :%s", obj)

		cmd := exec.Command("go", "build", "-o", obj)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		_, target, err := runner.RootFS().Open(runner.Current())

		if err != nil {
			return err
		}

		cmd.Dir = filepath.Join(target.Mapping, "cmd", "gsmake")

		return cmd.Run()
	}

	return nil
}
