package tasks

import (
	"flag"
	"fmt"
	"net/url"

	"github.com/gsdocker/gserrors"
	"github.com/gsmake/gsmake"
	"github.com/gsmake/gsmake/vfs"
)

func parseSCM(rootfs vfs.RootFS, ir *gsmake.Import) error {
	// calc scm url
	if ir.SCM == "" {

		u, err := url.Parse(fmt.Sprintf("https://%s", ir.Name))

		if err != nil {
			return gserrors.Newf(nil, "unknown host, must special -p flag")
		}

		ir.SCM = rootfs.Protocol(u.Host)
	}

	return nil
}

// TaskRedirect .
func TaskRedirect(runner *gsmake.Runner, args ...string) error {

	var flagset flag.FlagSet

	version := flagset.String("v", "current", "the package version")

	protocol := flagset.String("p", "", "the pacakge's scm protocol")

	if err := flagset.Parse(args); err != nil {
		return err
	}

	if flagset.NArg() != 1 {
		runner.I("usage : gsmake redirect [-v version] [-p protocol] package")
		return gserrors.Newf(nil, "expect redirect list package name")
	}

	if *protocol == "" {
		*protocol = runner.RootFS().Protocol(flagset.Arg(0))
	}

	src := fmt.Sprintf("%s://%s?version=%s", *protocol, flagset.Arg(0), *version)

	target := fmt.Sprintf("gsmake://%s?domain=redirect", flagset.Arg(0))

	runner.D("mount redirect config package:\n\tsrc :%s\n\ttarget :%s", src, target)

	err := runner.RootFS().Mount(src, target)

	if err != nil {
		return err
	}

	var redirect []struct {
		From gsmake.Import
		To   gsmake.Import
	}

	err = runner.Property("redirect", flagset.Arg(0), "redirect", &redirect)

	if err != nil {
		return err
	}

	for _, pair := range redirect {
		if err := parseSCM(runner.RootFS(), &pair.From); err != nil {
			return err
		}

		if err := parseSCM(runner.RootFS(), &pair.To); err != nil {
			return err
		}

		if pair.From.Version == "" {
			pair.From.Version = "current"
		}

		if pair.To.Version == "" {
			pair.To.Version = "current"
		}

		src := fmt.Sprintf("%s://%s?version=%s", pair.From.SCM, pair.From.Name, pair.From.Version)

		target := fmt.Sprintf("%s://%s?version=%s", pair.To.SCM, pair.To.Name, pair.To.Version)

		runner.I("redirect package :\n\tsrc :%s\n\ttarget :%s", src, target)

		if err := runner.RootFS().Redirect(src, target, true); err != nil {
			return err
		}
	}

	return nil
}
