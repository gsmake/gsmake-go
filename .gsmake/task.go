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

// TaskUpdate .
func TaskUpdate(context *gsmake.Runner, args ...string) error {

	var name string

	if len(args) != 0 {
		name = args[0]
	}

	if name == "" {
		fmt.Printf("%s package name :", context.Name())

		bio := bufio.NewReader(os.Stdin)
		line, _, err := bio.ReadLine()

		if err != nil {
			return gserrors.Newf(err, "read update package name error")
		}

		name = string(line)
	}

	return context.Update(name)
}

// TaskCompile .
func TaskCompile(context *gsmake.Runner, args ...string) error {

	var goinstall []string

	env, _ := exec.Command("go", "env").Output()

	context.V("go env\n%s", env)

	if context.PackageProperty(context.Name(), "goinstall", &goinstall) {

		for _, target := range goinstall {
			targetfile := filepath.Join(context.ResourceDir(), "bin", filepath.Base(target)+gsos.ExeSuffix)

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

			source := filepath.Join(context.ResourceDir(), "bin", name+gsos.ExeSuffix)

			target := filepath.Join(path, "bin", name+gsos.ExeSuffix)

			context.I("go install :\n\tfrom :%s\n\tto :%s", source, target)

			_, err := gsos.Copy(
				source,
				target,
				false,
			)

			if err != nil {
				return gserrors.Newf(err, "exec setup task error")
			}
		}
	}

	return nil
}

// TaskCache implement cache task
func TaskCache(context *gsmake.Runner, args ...string) error {

	return context.Cache()

}

// TaskList list loaded tasks
func TaskList(context *gsmake.Runner, args ...string) error {
	context.PrintTask()
	return nil
}

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
