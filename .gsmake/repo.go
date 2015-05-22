package plugin

import "github.com/gsmake/gsmake"

// TaskList list loaded tasks
func TaskList(context *gsmake.Runner, args ...string) error {
	context.PrintTask()
	return nil
}

// TaskUpdate .
func TaskUpdate(context *gsmake.Runner, args ...string) error {

	var name string

	if len(args) != 0 {
		name = args[0]
	}

	if name == "" {
		context.I("update all packages")
		return context.Repo().UpdateAll()
	}

	context.I("update package :%s", name)

	return context.Update(name)
}

// TaskUpdateall .
func TaskUpdateall(context *gsmake.Runner, args ...string) error {
	return context.Repo().UpdateAll()
}

// TaskCache implement cache task
func TaskCache(runner *gsmake.Runner, args ...string) error {
	return runner.Cache()
}

// TaskRmcache implement cache task
func TaskRmcache(runner *gsmake.Runner, args ...string) error {
	return runner.RemoveCache()
}
