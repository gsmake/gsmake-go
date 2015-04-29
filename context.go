package gsmake

import (
	"errors"
	"fmt"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
)

// errors
var (
	ErrTask = errors.New("task error")
)

// Task gsmake task
type Task struct {
	Name string // task name
	F    TaskF  // task function
	Prev string // prev task name
}

// TaskF task function
type TaskF func(context *Context) error

// Context .
type Context struct {
	gslogger.Log                    // mixin log APIs
	name         string             // project Name
	version      string             // project version
	path         string             // project path
	root         string             // gsmake root path
	current      *Task              // current execute task
	tasks        map[string][]*Task // register tasks
}

// NewContext .
func NewContext() *Context {
	return &Context{
		Log:   gslogger.Get("gsmake"),
		tasks: make(map[string][]*Task),
	}
}

// Init .
func (context *Context) Init(name string, version string, path string, root string) {
	context.name = name
	context.version = version
	context.path = path
	context.root = root
}

// Task register task
func (context *Context) Task(task *Task) {
	context.tasks[task.Name] = append(context.tasks[task.Name], task)
}

//ListTask .
func (context *Context) ListTask() {
	fmt.Println("register task :")
	for name := range context.tasks {
		fmt.Printf("* %s\n", name)
	}
}

// RunTask run task
func (context *Context) RunTask(name string) error {

	if taskQ, ok := context.tasks[name]; ok {

		for _, task := range taskQ {

		}

		return nil
	}

	return gserrors.Newf(ErrTask, "unregister task :%s", name)
}
