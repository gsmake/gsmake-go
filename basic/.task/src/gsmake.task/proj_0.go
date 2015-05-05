package main

import "github.com/gsdocker/gsmake"
import task "github.com/gsdocker/gsmake/.gsmake"

func init() {

	context.Task(&gsmake.TaskCmd{
		Name:    "compile",
		F:       task.TaskCompile,
		Prev:    "resources",
		Project: "",
	})

	context.Task(&gsmake.TaskCmd{
		Name:    "publish",
		F:       task.TaskPublish,
		Prev:    "test",
		Project: "",
	})

	context.Task(&gsmake.TaskCmd{
		Name:    "resources",
		F:       task.TaskResources,
		Prev:    "",
		Project: "",
	})

	context.Task(&gsmake.TaskCmd{
		Name:    "test",
		F:       task.TaskTest,
		Prev:    "compile",
		Project: "",
	})

}
