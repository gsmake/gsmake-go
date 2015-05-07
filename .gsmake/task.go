package plugin

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gsmake"
	"github.com/gsdocker/gsos"
)

// TaskResources .
func TaskResources(context *gsmake.Runner, args ...string) error {
	context.D("TODO: invoke gslang tools")
	return nil
}

// TaskCompile .
func TaskCompile(context *gsmake.Runner, args ...string) error {

	var goinstall []string

	env, _ := exec.Command("go", "env").Output()

	context.V("go env\n%s", env)

	if context.PackageProperty(context.Name(), "goinstall", &goinstall) {

		for _, target := range goinstall {
			targetfile := filepath.Join(context.ResourceDir(), "bin", filepath.Base(target))

			context.I("[gocompiler] generate target :%s", filepath.Base(target))

			cmd := exec.Command("go", "build", "-o", targetfile, filepath.Join(context.Name(), target))

			cmd.Stdin = os.Stdin

			cmd.Stdout = os.Stdout

			err := cmd.Run()

			if err != nil {
				return err
			}
		}
	}

	//
	return nil
}

// TaskTest .
func TaskTest(context *gsmake.Runner, args ...string) error {
	context.D("hello test")
	return nil
}

// TaskSetup .
func TaskSetup(context *gsmake.Runner, args ...string) error {

	var path string

	if len(args) != 0 {
		path = args[0]
	}

	if path == "" {
		fmt.Printf("%s install path :", context.Name())

		bio := bufio.NewReader(os.Stdin)
		line, _, err := bio.ReadLine()

		if err != nil {
			return gserrors.Newf(err, "read install path error")
		}

		path = string(line)
	}

	if err := os.MkdirAll(filepath.Join(path, "bin"), 0755); err != nil {
		return gserrors.Newf(err, "create setup target directory error")
	}

	var goinstall []string

	if context.PackageProperty(context.Name(), "goinstall", &goinstall) {
		for _, target := range goinstall {

			name := filepath.Base(target)

			_, err := gsos.Copy(
				filepath.Join(context.ResourceDir(), "bin", name),
				filepath.Join(path, "bin", name),
				false,
			)

			if err != nil {
				return gserrors.Newf(err, "exec setup task error")
			}
		}
	}

	return nil
}

// TaskList list loaded tasks
func TaskList(context *gsmake.Runner, args ...string) error {
	context.PrintTask()
	return nil
}

// TaskDev setup package develop enverioment
func TaskDev(context *gsmake.Runner, args ...string) error {

	if len(args) == 0 {
		context.E("expect package name")
		return nil
	}

	linktarget := filepath.Join(context.StartDir(), filepath.Base(args[0]))

	if gsos.IsExist(linktarget) {

		context.E("package directory already exists\n\t%s", linktarget)

		return nil
	}

	err := context.Link(args[0], "current", linktarget)

	if err != nil {
		return err
	}

	context.I("setup package [%s] develop environment -- success", args[0])

	return nil
}

// TaskAtom setup package develop enverioment
func TaskAtom(context *gsmake.Runner, args ...string) error {

	context.I("start atom ...")

	cmd := exec.Command("atom", context.StartDir())

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
