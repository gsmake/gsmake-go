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

	if runner.Name() == "github.com/gsmake/gsmake" {

		if len(args) == 0 {
			return gserrors.Newf(nil, "expect setup dir")
		}

		os.Setenv("GOPATH", runner.RootFS().DomainDir("task"))

		obj := filepath.Join(args[0], "bin", "gsmake"+gsos.ExeSuffix)

		runner.I("install gsmake to :%s", obj)

		cmd := exec.Command("go", "build", "-o", obj)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		target, err := runner.Path("task", "github.com/gsmake/gsmake")

		if err != nil {
			return err
		}

		cmd.Dir = filepath.Join(target, "cmd", "gsmake")

		return cmd.Run()
	}

	return nil
}
