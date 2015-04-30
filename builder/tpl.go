package builder

var tpl = `

{{define "main.go"}}
// generate builder for {{.Name}}

package main

import "flag"
import "os"
import "github.com/gsdocker/gslogger"
import "github.com/gsdocker/gsmake"

var context = gsmake.NewContext()

var listask = flag.Bool("task",false,"list all register task")

func main(){
    defer gslogger.Join()

    context.Init("{{.Name}}","current","{{ospath .Path}}","{{ospath .Root}}")

    flag.Parse()



    if *listask {
        context.ListTask()
        return
    }

    if flag.NArg() != 1 {
        flag.PrintDefaults()
        os.Exit(1)
    }

    if err := context.RunTask(flag.Arg(0)); err != nil {
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
    context.Task(&gsmake.Task{
        Name : "{{$key}}",
        F : task.{{taskname $key}},
        Prev : "{{$value.Dependency}}",
        Project : "{{$value.Project.Name}}",
    })
    {{end}}
}

{{end}}
`
