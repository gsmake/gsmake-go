package tasks

import (
	"fmt"

	"github.com/gsdocker/gserrors"
	"github.com/gsmake/gsmake"
)

// TaskCache .
func TaskCache(runner *gsmake.Runner, args ...string) error {

	if len(args) != 1 {
		return gserrors.Newf(nil, "usage : gsmake cache ${url}")
	}

	target := fmt.Sprintf("file://%s", runner.StartDir())

	return runner.RootFS().Redirect(args[0], target, true)
}

// TaskDiscache .
func TaskDiscache(runner *gsmake.Runner, args ...string) error {
	if len(args) != 1 {
		return gserrors.Newf(nil, "usage : gsmake discache ${url}")
	}

	target := fmt.Sprintf("file://%s", runner.StartDir())

	return runner.RootFS().Redirect(args[0], target, false)
}
