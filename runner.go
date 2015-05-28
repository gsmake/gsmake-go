package gsmake

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
	"github.com/gsmake/gsmake/vfs"
)

// ErrTask .
var (
	ErrTask = errors.New("error task")
)

// TaskCmd gsmake task
type TaskCmd struct {
	Name        string // task name
	Description string // task description
	F           TaskF  // task function
	Prev        string // prev task name
	Project     string // project belongs to
}

// TaskF task function
type TaskF func(runner *Runner, args ...string) error

type visitMark int

const (
	white visitMark = iota // has not been visited
	gray                   // visiting
	black                  // visited
)

type taskGroup struct {
	name        string     // task name
	description string     // description
	group       []*TaskCmd // group slice
	mark        visitMark  // visit mark
}

func (group *taskGroup) add(task *TaskCmd) {
	group.group = append(group.group, task)

	if group.description == "" {
		group.description = task.Description
	}
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

func (group *taskGroup) invoke(runner *Runner, args ...string) error {

	for _, task := range group.group {
		if err := task.F(runner, args...); err != nil {
			return err
		}
	}

	return nil
}

// Runner gsmake task runner
type Runner struct {
	gslogger.Log                       // mixin Logger
	current      *Task                 // current execute task
	tasks        map[string]*taskGroup // register tasks
	checkerOfDCG []*taskGroup          // DCG check stack
	rootfs       vfs.RootFS            // rootfs
	rootpath     string                // gsmake root path
	targetpath   string                // the processing root package path
	currentpkg   *Package              // current handle package object
}

// NewRunner create new task runner
func NewRunner(rootpath string, targetpath string) *Runner {

	runner := &Runner{
		Log:        gslogger.Get("gsmake"),
		tasks:      make(map[string]*taskGroup),
		rootpath:   rootpath,
		targetpath: targetpath,
	}

	return runner
}

// Current gsmake current handle pacakge
func (runner *Runner) Current() string {
	return fmt.Sprintf("gsmake://%s?domain=task", runner.currentpkg.Name)
}

// RootFS get rootfs object
func (runner *Runner) RootFS() vfs.RootFS {
	return runner.rootfs
}

// Start .
func (runner *Runner) Start() error {
	rootfs, err := vfs.New(runner.rootpath, runner.targetpath)

	if err != nil {
		return err
	}

	runner.rootfs = rootfs

	jsonfile := filepath.Join(runner.targetpath, ".gsmake.json")

	runner.currentpkg, err = loadjson(jsonfile)

	if err != nil {
		return err
	}

	return nil
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
	var stream bytes.Buffer
	stream.WriteString("task list:\n")
	for name, task := range runner.tasks {
		stream.WriteString(fmt.Sprintf("\t* %s%s;%s\n", name, strings.Repeat(" ", 20-len(name)), task.description))
	}

	runner.I("print tasks\n%s", stream.String())
}

func (runner *Runner) unmark() {
	for _, group := range runner.tasks {
		group.unmark()
	}
}

// Run run task
func (runner *Runner) Run(name string, args ...string) error {

	//DFS Topo sort

	if group, ok := runner.tasks[name]; ok {

		result, err := group.topoShort(runner)

		runner.unmark()

		if err != nil {
			return err
		}

		for _, group := range result {
			if err := group.invoke(runner, args...); err != nil {
				return err
			}
		}

		return nil
	}

	return gserrors.Newf(ErrTask, "unknown task :%s", name)
}
