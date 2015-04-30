package gsmake

import (
	"bytes"
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
	Name    string // task name
	F       TaskF  // task function
	Prev    string // prev task name
	Project string // project belongs to
}

// TaskF task function
type TaskF func(context *Context) error

type visitMark int

const (
	white visitMark = iota // has not been visited
	gray                   // visiting
	black                  // visited
)

type taskGroup struct {
	name  string    // task name
	group []*Task   // group slice
	mark  visitMark // visit mark
}

func (group *taskGroup) add(task *Task) {
	group.group = append(group.group, task)
}

func (group *taskGroup) unmark() {
	group.mark = white
}

func (group *taskGroup) topoShort(context *Context) ([]*taskGroup, error) {

	if group.mark == black {
		return nil, nil
	}

	//DCG check
	if group.mark == gray {

		var stream bytes.Buffer

		stream.WriteString(fmt.Sprintf("%s -> ", group.name))

		for i := len(context.checkerOfDCG) - 1; i > 0; i-- {
			stream.WriteString(fmt.Sprintf("%s ->", context.checkerOfDCG[i].name))

			if context.checkerOfDCG[i] == group {
				break
			}
		}

		return nil, gserrors.Newf(ErrTask, "DCG detected : %s", stream.String())
	}

	group.mark = gray
	context.checkerOfDCG = append(context.checkerOfDCG, group)

	defer func() {
		context.checkerOfDCG = context.checkerOfDCG[:len(context.checkerOfDCG)-1]
	}()

	var result []*taskGroup

	for _, task := range group.group {

		if task.Prev == "" {
			continue
		}

		if prev, ok := context.tasks[task.Prev]; ok {
			r, err := prev.topoShort(context)

			if err != nil {
				return nil, err
			}

			result = append(result, r...)
		} else {
			return nil, gserrors.Newf(ErrTask, "unknown task %s which is reference by %s:%s", task.Prev, task.Project, task.Name)
		}
	}

	group.mark = black

	result = append(result, group)

	return result, nil
}

func (group *taskGroup) invoke(context *Context) error {

	for _, task := range group.group {
		if err := task.F(context); err != nil {
			return err
		}
	}

	return nil
}

// Context .
type Context struct {
	gslogger.Log                       // mixin log APIs
	name         string                // project Name
	version      string                // project version
	path         string                // project path
	root         string                // gsmake root path
	current      *Task                 // current execute task
	tasks        map[string]*taskGroup // register tasks
	checkerOfDCG []*taskGroup          // DCG check stack
}

// NewContext .
func NewContext() *Context {
	return &Context{
		Log:   gslogger.Get("gsmake"),
		tasks: make(map[string]*taskGroup),
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
	group, ok := context.tasks[task.Name]
	if !ok {
		group = &taskGroup{name: task.Name}
		context.tasks[task.Name] = group
	}

	group.add(task)
}

//ListTask .
func (context *Context) ListTask() {
	fmt.Println("register task :")
	for name := range context.tasks {
		fmt.Printf("* %s\n", name)
	}
}

func (context *Context) unmark() {
	for _, group := range context.tasks {
		group.unmark()
	}
}

// RunTask run task
func (context *Context) RunTask(name string) error {

	//DFS Topo sort

	if group, ok := context.tasks[name]; ok {

		result, err := group.topoShort(context)

		context.unmark()

		if err != nil {
			return err
		}

		for _, group := range result {
			if err := group.invoke(context); err != nil {
				return err
			}
		}

		return nil
	}

	return gserrors.Newf(ErrTask, "unregister task :%s", name)
}
