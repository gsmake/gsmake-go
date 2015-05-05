package gsmake

var codegen = `
{{define "main.go"}}
// generate builder for {{.Name}}

package main

import "flag"
import "os"
import "github.com/gsdocker/gslogger"
import "github.com/gsdocker/gsmake"

var context = gsmake.NewRunner("{{.Name}}","{{ospath .Path}}","{{ospath .Root}}")

var listask = flag.Bool("task",false,"list all register task")

func main(){
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
        context.E("%s",err)
        gslogger.Join()
        os.Exit(1)
    }
}

{{end}}


{{define "project.go"}}
package main
import "github.com/gsdocker/gsmake"
import task "{{.Name}}/.gsmake"

func init(){

    {{range $key, $value := .Task}}
    context.Task(&gsmake.TaskCmd{
        Name : "{{$key}}",
        F : task.{{taskname $key}},
        Prev : "{{$value.Prev}}",
        Project : "{{$value.Package}}",
    })
    {{end}}
}

{{end}}
`
