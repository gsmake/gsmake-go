package tasks

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gsos/fs"
	"github.com/gsmake/gsmake"
)

// TaskCreate create package base on archtype
func TaskCreate(runner *gsmake.Runner, args ...string) error {

	if runner.Name() != gsmake.PacakgeAnonymous {
		return gserrors.Newf(nil, "you are already in a package dir.")
	}

	var flagset flag.FlagSet

	output := flagset.String("o", "", "the package version")

	version := flagset.String("v", "current", "the package version")

	protocol := flagset.String("p", "", "the pacakge's scm protocol")

	if err := flagset.Parse(args); err != nil {
		return err
	}

	if flagset.NArg() != 1 {
		runner.I("usage : gsmake create [-v version] [-p protocol] package:archtype")
		return gserrors.Newf(nil, "expect package:archtype arg")
	}

	token := strings.SplitN(flagset.Arg(0), ":", 2)

	if len(token) != 2 {
		runner.I("usage : gsmake create [-v version] [-p protocol] package:archtype")
		return gserrors.Newf(nil, "invalid arg :%s", flagset.Arg(0))
	}

	host := token[0]

	name := token[1]

	if *output == "" {
		*output = filepath.Join(runner.StartDir(), filepath.Base(name))
	} else {
		*output, _ = filepath.Abs(*output)
	}

	if *protocol == "" {
		*protocol = runner.RootFS().Protocol(host)
	}

	runner.I("package :%s", host)
	runner.I("archtype :%s", name)
	runner.I("protocol :%s", *protocol)
	runner.I("target :%s", *output)

	src := fmt.Sprintf("%s://%s?version=%s", *protocol, host, *version)

	target := fmt.Sprintf("gsmake://%s?domain=archtype", host)

	err := runner.RootFS().Mount(src, target)

	if err != nil {
		return err
	}

	path, err := runner.Path("task", host)

	if err != nil {
		return err
	}

	archtype := filepath.Join(path, ".archtype", name)

	if !fs.Exists(archtype) {
		return gserrors.Newf(nil, "archtype not exists :", flagset.Arg(0))
	}

	if fs.Exists(*output) {
		return gserrors.Newf(nil, "target dir already exists")
	}

	if err := fs.CopyDir(archtype, *output); err != nil {
		return gserrors.Newf(err, "copy archtype to target dir error")
	}

	return nil

}
