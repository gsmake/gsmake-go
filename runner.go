package gsmake

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gsdocker/gsconfig"
	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
	"github.com/gsdocker/gsos"
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
	root         string                // gsmake root path
	path         string                // current running package path
	name         string                // current running package name
	rcdir        string                // runtime resources dir
	rundir       string                // start running directory
	current      *Task                 // current execute task
	tasks        map[string]*taskGroup // register tasks
	checkerOfDCG []*taskGroup          // DCG check stack
	packages     map[string]*Package   //  loaded packages
	repository   *Repository           // Repository
}

// NewRunner create new task runner
func NewRunner(name string, path string, root string) *Runner {

	return &Runner{
		Log:    gslogger.Get("gsmake"),
		root:   root,
		path:   path,
		rundir: gsos.CurrentDir(),
		name:   name,
		rcdir:  filepath.Join(path, gsconfig.String("gsmake.rundir", ".run")),
		tasks:  make(map[string]*taskGroup),
	}
}

// Package query package by name
func (runner *Runner) Package(name string) (pkg *Package, ok bool) {
	pkg, ok = runner.packages[name]
	return
}

// Packages loop loade packages
func (runner *Runner) Packages(f func(*Package) bool) {
	for _, pkg := range runner.packages {
		if !f(pkg) {
			return
		}
	}
}

// StartDir task runner start directory
func (runner *Runner) StartDir() string {
	return runner.rundir
}

// PackageProperty get package property by name
func (runner *Runner) PackageProperty(name string, key string, target interface{}) bool {

	if pkg, ok := runner.Package(name); ok {
		if v, ok := pkg.Properties[key]; ok {
			content, _ := json.Marshal(v)

			json.Unmarshal(content, target)

			return true
		}
	}

	return false
}

// Name current package name
func (runner *Runner) Name() string {
	return runner.name
}

// ResourceDir runtime resource dir
func (runner *Runner) ResourceDir() string {
	return runner.rcdir
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

	loader, err := Load(runner.root, runner.path, stageRuntimes)

	if err != nil {
		return err
	}

	runner.packages = loader.packages
	runner.repository = loader.repository

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
