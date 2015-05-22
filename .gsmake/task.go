package plugin

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gsos"
	"github.com/gsmake/gsmake"
)

// TaskResources .
func TaskResources(context *gsmake.Runner, args ...string) error {
	return nil
}

// TaskCompile .
func TaskCompile(context *gsmake.Runner, args ...string) error {

	var goinstall []string

	env, _ := exec.Command("go", "env").Output()

	context.V("go env\n%s", env)

	if context.PackageProperty(context.Name(), "goinstall", &goinstall) {

		for _, target := range goinstall {
			targetfile := filepath.Join(context.Workspace(), "bin", filepath.Base(target)+gsos.ExeSuffix)

			context.I("[gocompiler] generate target :%s", filepath.Base(target))

			cmd := exec.Command("go", "build", "-o", targetfile, filepath.Join(context.Name(), target))

			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

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

			source := filepath.Join(context.Workspace(), "bin", name+gsos.ExeSuffix)

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

// TaskGotest implement gotest task
func TaskGotest(context *gsmake.Runner, args ...string) error {

	err := os.Chdir(context.StartDir())

	if err != nil {
		return err
	}

	newargs := []string{"test"}

	newargs = append(newargs, args...)

	cmd := exec.Command("go", newargs...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
