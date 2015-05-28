package tasks

import "github.com/gsmake/gsmake"

// TaskList .
func TaskList(runner *gsmake.Runner, args ...string) error {
	runner.PrintTask()
	return nil
}
