package builder

var tpl = `

{{define "main.go"}}
// generate builder for {{.RootProject}}

package main

import "flag"
import "github.com/gsdocker/gslogger"
import "github.com/gsdocker/gsmake"

var context = gsmake.NewContext()

var listask = flag.Bool("task",false,"list all register task")

func main(){
    defer gslogger.Join()

    context.Init("{{.RootProject}}","current","{{.Path}}","{{.Root}}")

    flag.Parse()

    if *listask {
        context.ListTask()
        return
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
    })
    {{end}}
}

{{end}}
`
