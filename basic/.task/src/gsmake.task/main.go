// generate builder for github.com/gsdocker/gsmake/basic

package main

import "flag"
import "os"
import "github.com/gsdocker/gslogger"
import "github.com/gsdocker/gsmake"

var context = gsmake.NewRunner("github.com/gsdocker/gsmake/basic", "/Users/liyang/Workspace/go/src/github.com/gsdocker/gsmake/basic", "/Users/liyang/Workspace/go/src/github.com/gsdocker/gsmake/.repo")

var listask = flag.Bool("task", false, "list all register task")

func main() {
	defer gslogger.Join()

	flag.Parse()

	if *listask {
		context.PrintTask()
		return
	}

	if flag.NArg() != 1 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err := context.Run(flag.Arg(0)); err != nil {
		context.E("%s", err)
		gslogger.Join()
		os.Exit(1)
	}
}
