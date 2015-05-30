package tasks

import (
	"flag"

	"github.com/gsmake/gsmake"
)

// TaskUpdate .
func TaskUpdate(runner *gsmake.Runner, args ...string) error {
	var flagSet flag.FlagSet

	nocache := flagSet.Bool("nocache", false, "also update cached package")

	if err := flagSet.Parse(args); err != nil {
		return err
	}
	
	if runner.Name() == gsmake.PacakgeAnonymous {
		*nocache = true
	}

	return runner.RootFS().UpdateAll(*nocache)
}
