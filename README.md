# gsmake

gsmake is a build automation software which is develop by pure go language

## setup

> 1. install golang sdk
> 2. install git
> 3. go get github.com/gsdocker/gsmake
> 4. cd {gsmake source dir}
> 5. run ./setup {install dir}
> 6. append {install dir}/bin to system env $PATH

## usage

### create customer gsmake task

1. create project dir sample
2. cd sample
3. create .gsmake.json config file in project dir:
```json
{
    "name":"github.com/gsdocker/sample",

    "import" : [
        {"name" : "github.com/gsmake/gsmake"}
    ],

    "task" : {

        "helloworld" : {"description" : "say hello"},
    }
}
```
4. create go source file : {project.dir}/.gsmake/task.go:
```go
package plugin

import "github.com/gsmake/gsmake"

// TaskHelloworld implement gsmake task helloworld .
func TaskHelloworld(context *gsmake.Runner, args ...string) error {
	context.I("hello gsmake!!!!!!")
	return nil
}
```
5. run task helloworld
> gsmake helloworld


### golang project build

> the gsmake project is a good example for build golang project ,see ./gsmake.json for more detail
