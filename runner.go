package gsmake

import (
	"bytes"
	"fmt"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
)

// TaskCmd gsmake task
type TaskCmd struct {
	Name    string // task name
	F       TaskF  // task function
	Prev    string // prev task name
	Project string // project belongs to
}

// TaskF task function
type TaskF func(runner *Runner) error

type visitMark int

const (
	white visitMark = iota // has not been visited
	gray                   // visiting
	black                  // visited
)

type taskGroup struct {
	name  string     // task name
	group []*TaskCmd // group slice
	mark  visitMark  // visit mark
}

func (group *taskGroup) add(task *TaskCmd) {
	group.group = append(group.group, task)
}

func (group *taskGroup) unmark() {
	group.mark = white
}

func (group *taskGroup) topoShort(context *Runner) ([]*taskGroup, error) {

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

func (group *taskGroup) invoke(runner *Runner) error {

	for _, task := range group.group {
		if err := task.F(runner); err != nil {
			return err
		}
	}

	return nil
}

// Runner gsmake task runner
type Runner struct {
	gslogger.Log                       // mixin Logger
	root         string                // gsmake root path
	path         string                // current running package path
	name         string                // current running package name
	current      *Task                 // current execute task
	tasks        map[string]*taskGroup // register tasks
	checkerOfDCG []*taskGroup          // DCG check stack
}

// NewRunner create new task runner
func NewRunner(root string, path string, name string) *Runner {
	return &Runner{
		Log:   gslogger.Get("gsmake"),
		root:  root,
		path:  path,
		name:  name,
		tasks: make(map[string]*taskGroup),
	}
}

// Task register task
func (runner *Runner) Task(task *TaskCmd) {
	group, ok := runner.tasks[task.Name]
	if !ok {
		group = &taskGroup{name: task.Name}
		runner.tasks[task.Name] = group
	}

	group.add(task)
}

// PrintTask print defined task list
func (runner *Runner) PrintTask() {
	fmt.Println("register task :")
	for name := range runner.tasks {
		fmt.Printf("* %s\n", name)
	}
}

func (runner *Runner) unmark() {
	for _, group := range runner.tasks {
		group.unmark()
	}
}

// Run run task
func (runner *Runner) Run(name string) error {

	//DFS Topo sort

	if group, ok := runner.tasks[name]; ok {

		result, err := group.topoShort(runner)

		runner.unmark()

		if err != nil {
			return err
		}

		for _, group := range result {
			if err := group.invoke(runner); err != nil {
				return err
			}
		}

		return nil
	}

	return gserrors.Newf(ErrTask, "unregister task :%s", name)
}
