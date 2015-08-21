package tasks

import (
	"flag"
	"fmt"

	"github.com/gsmake/gsmake"
)

// TaskCache .
func TaskCache(runner *gsmake.Runner, args ...string) error {

	var flagSet flag.FlagSet

	version := flagSet.String("v", "current", "package version")

	protocol := flagSet.String("p", runner.SCM(), "scm protocol")

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	src := fmt.Sprintf("%s://%s?version=%s", *protocol, runner.Name(), *version)

	target := fmt.Sprintf("file://%s", runner.StartDir())

	runner.I("cache package :\n\tsrc :%s\n\ttarget :%s", src, target)

	return runner.RootFS().Redirect(src, target, true)
}

// TaskDiscache .
func TaskDiscache(runner *gsmake.Runner, args ...string) error {
	var flagSet flag.FlagSet

	version := flagSet.String("v", "current", "package version")

	protocol := flagSet.String("p", runner.SCM(), "scm protocol")

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	src := fmt.Sprintf("%s://%s?version=%s", *protocol, runner.Name(), *version)

	target := fmt.Sprintf("file://%s", runner.StartDir())

	runner.I("discache package :\n\tsrc :%s\n\ttarget :%s", src, target)

	return runner.RootFS().Redirect(src, target, false)
}
