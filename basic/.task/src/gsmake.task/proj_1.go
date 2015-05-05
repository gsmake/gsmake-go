package main

import "github.com/gsdocker/gsmake"
import task "github.com/gsdocker/gsmake/basic/.gsmake"

func init() {

	context.Task(&gsmake.TaskCmd{
		Name:    "compile",
		F:       task.TaskCompile,
		Prev:    "gs2go",
		Project: "",
	})

	context.Task(&gsmake.TaskCmd{
		Name:    "gs2go",
		F:       task.TaskGs2go,
		Prev:    "resources",
		Project: "",
	})

}
