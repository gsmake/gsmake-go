package main

import "github.com/gsmake/gsmake"
import task "github.com/gsmake/gsmake/.gsmake"

func init() {

	context.Task(&gsmake.TaskCmd{
		Name:        "list",
		Description: "list tasks",
		F:           task.TaskList,
		Prev:        "",
		Project:     "",
	})

	context.Task(&gsmake.TaskCmd{
		Name:        "setup",
		Description: "install current package",
		F:           task.TaskSetup,
		Prev:        "",
		Project:     "",
	})

}
