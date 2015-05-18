package plugin

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gsmake"
	"github.com/gsdocker/gsos"
	"github.com/gsdocker/gsos/uuid"
)

// TaskArchetype implement task archetype
func TaskArchetype(runner *gsmake.Runner, args ...string) error {
	flagset := flag.NewFlagSet("archetype", flag.ContinueOnError)

	var (
		name    = flagset.String("p", "", "archetype provide package name")
		version = flagset.String("v", "", "archetype provide package version")
	)

	if err := flagset.Parse(args); err != nil {
		return err
	}

	if flagset.NFlag() != 2 && flagset.NArg() != 2 {
		return gserrors.Newf(nil, "usage : gsmake archetype -p {name} -v {version} {archetype name} {new package name}")
	}

	path := filepath.Join(os.TempDir(), uuid.New())

	err := runner.Repo().Copy(*name, *version, path)

	if err != nil {
		return err
	}

	path = filepath.Join(path, ".archtype", flagset.Arg(0))

	if !gsos.IsDir(path) {
		return gserrors.Newf(nil, "archetype(%s) not exist", flagset.Arg(0))
	}

	target := filepath.Join(runner.StartDir(), filepath.Base(flagset.Arg(1)))

	if err := gsos.CopyDir(path, target); err != nil {
		return err
	}

	jsonfile := filepath.Join(target, ".gsmake.json")

	if gsos.IsExist(jsonfile) {
		content, err := ioutil.ReadFile(jsonfile)

		if err != nil {
			return err
		}

		properties := gsmake.Properties{
			"name": flagset.Arg(1),
		}

		content = []byte(gsmake.Expand(string(content), properties))

		return ioutil.WriteFile(jsonfile, content, 0644)
	}

	return nil
}
