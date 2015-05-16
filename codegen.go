package gsmake

var codegen = `
{{define "main.go"}}
// generate builder for {{.Name}}

package main

import "os"
import "fmt"
import "flag"
import "strings"
import "github.com/gsdocker/gslogger"
import "github.com/gsdocker/gsmake"

var cacheflag = flag.Bool("nocache", false, "using caching packages")
var verbflag = flag.Bool("v", false, "print more debug information")

var context = gsmake.NewRunner("{{.Name}}","{{ospath .Path}}","{{ospath .Root}}")


func main(){

    flag.Parse()

    if flag.NArg() < 1 {
        fmt.Println("expect task name")
        os.Exit(1)
    }

    if !*verbflag {
		gslogger.NewFlags(gslogger.ASSERT | gslogger.ERROR | gslogger.WARN | gslogger.INFO)
	}

    if err := context.Start(*cacheflag); err != nil {
        context.E("%s",err)
        gslogger.Join()
        os.Exit(1)
    }

    context.I("exec task [%s] with args : %s",flag.Arg(0),strings.Join(flag.Args()[1:]," "))

    if err := context.Run(flag.Arg(0),flag.Args()[1:]...); err != nil {
        context.E("%s",err)
        gslogger.Join()
        os.Exit(1)
    }

    gslogger.Join()
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
        Description : "{{$value.Description}}",
        F : task.{{taskname $key}},
        Prev : "{{$value.Prev}}",
        Project : "{{$value.Package}}",
    })
    {{end}}
}

{{end}}
`
