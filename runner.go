package gsmake

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

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
	homepath     string                // gsmake root path
	workspace    string                // runtime resources dir
	name         string                // current running package name
	path         string                // root package path
	startdir     string                // start running directory
	current      *Task                 // current execute task
	tasks        map[string]*taskGroup // register tasks
	checkerOfDCG []*taskGroup          // DCG check stack
	repository   *Repository           // Repository
	loader       *Loader               // package loader
}

// NewRunner create new task runner
func NewRunner(name string, path string, homepath string) *Runner {
	runner := &Runner{
		Log:       gslogger.Get("gsmake"),
		name:      name,
		path:      path,
		homepath:  homepath,
		startdir:  gsos.CurrentDir(),
		workspace: Workspace(homepath, name),
		tasks:     make(map[string]*taskGroup),
	}

	return runner
}

// LoadPackage .
func (runner *Runner) LoadPackage(name string, version string) (string, error) {
	pkg, err := runner.loader.LoadPackage(name, version)

	if err != nil {
		return "", err
	}

	return pkg.origin, nil
}

// Start .
func (runner *Runner) Start(nocached bool) error {
	loader, err := Load(runner.homepath, runner.path, stageRuntimes, nocached, nil)

	if err != nil {
		return err
	}

	runner.repository = loader.repository
	runner.loader = loader
	return nil
}

// Package query package by name
func (runner *Runner) Package(name string) (pkg *Package, ok bool) {
	pkg, ok = runner.loader.packages[name]
	return
}

// Update update current package
func (runner *Runner) Update(name string) error {
	_, err := runner.repository.Update(name, "current")

	return err
}

// Packages loop loaded packages
func (runner *Runner) Packages(f func(*Package) bool) {
	for _, pkg := range runner.loader.packages {
		if !f(pkg) {
			return
		}
	}
}

// StartDir task runner start directory
func (runner *Runner) StartDir() string {
	return runner.startdir
}

// Repo .
func (runner *Runner) Repo() *Repository {
	return runner.repository
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

// Workspace runtime Workspace dir
func (runner *Runner) Workspace() string {
	return runner.workspace
}

// Home gsmake home path
func (runner *Runner) Home() string {
	return runner.homepath
}

// Cache link develop package into gsmake cache space
func (runner *Runner) Cache() error {

	if !gsos.IsExist(filepath.Join(runner.startdir, ".gsmake.json")) {
		return gserrors.Newf(nil, "expect .gsmake.json file")
	}

	return runner.repository.Cache(runner.name, "current", runner.StartDir())
}

// RemoveCache .
func (runner *Runner) RemoveCache() error {

	if !gsos.IsExist(filepath.Join(runner.startdir, ".gsmake.json")) {
		return gserrors.Newf(nil, "expect .gsmake.json file")
	}

	return runner.repository.RemoveCache(runner.name, "current", runner.StartDir())
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
