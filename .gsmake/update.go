package tasks

import (
	"flag"
	"fmt"

	"github.com/gsmake/gsmake"
	"github.com/gsmake/gsmake/vfs"
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

	if flagSet.NArg() > 0 {

		if runner.Name() == gsmake.PacakgeAnonymous {
			for _, target := range flagSet.Args() {
				if err := runner.RootFS().UpdateCache(fmt.Sprintf("gsmake://%s", target)); err != nil {
					return err
				}
			}

		} else {
			for _, target := range flagSet.Args() {

				err := runner.RootFS().List(func(srcE, targetE *vfs.Entry) bool {

					if fmt.Sprintf("%s%s", targetE.Host, targetE.Path) == target {
						err := runner.RootFS().Update(targetE.String(), *nocache)

						if err != nil {
							panic(err)
						}
					}

					return true
				})

				if err != nil {
					return err
				}
			}
		}

		return nil
	}

	return runner.RootFS().UpdateAll(*nocache)
}
