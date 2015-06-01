package gsmake

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
	"github.com/gsdocker/gsos/fs"
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
	Scope       string // scope belongs to
}

func (cmd *TaskCmd) String() string {

	scope := strings.ToUpper(cmd.Scope)

	if scope == "" {
		scope = "ALL"
	}

	return fmt.Sprintf(
		"task details\n\tpackage:     %s\n\tdescription: %s\n\tscope:       %s\n",
		cmd.Name,
		cmd.Description,
		scope,
	)
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

func (group *taskGroup) invoke(runner *Runner, domain string, args ...string) error {

	for _, task := range group.group {

		scope := task.Scope

		if scope != "" {

			scopes := strings.Split(scope, "|")

			skip := true

			for _, v := range scopes {
				if v == domain {
					skip = false
					break
				}

				if v == "all" {
					skip = false
					break
				}
			}

			if skip {

				runner.I("skip task\n%s", task)

				continue
			}
		} else {
			scope = "ALL"
		}

		runner.I("exec task ...\n%s", task)

		startime := time.Now()

		if err := task.F(runner, args...); err != nil {
			return err
		}

		runner.I("exec task -- success %s", time.Now().Sub(startime))

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
	startdir     string                // runner start dir
}

// NewRunner create new task runner
func NewRunner(rootpath string, targetpath string) *Runner {

	runner := &Runner{
		Log:        gslogger.Get("gsmake"),
		tasks:      make(map[string]*taskGroup),
		rootpath:   rootpath,
		targetpath: targetpath,
		startdir:   fs.Current(),
	}

	return runner
}

// Name get package name
func (runner *Runner) Name() string {
	return runner.currentpkg.Name
}

// Path get package's path
func (runner *Runner) Path(domain, name string) (string, error) {
	_, target, err := runner.rootfs.Open(fmt.Sprintf("gsmake://%s?domain=%s", name, domain))

	if err != nil {
		return "", err
	}

	return target.Mapping, nil
}

// Property get package's property
func (runner *Runner) Property(domain, packagename, name string, val interface{}) error {

	_, target, err := runner.rootfs.Open(fmt.Sprintf("gsmake://%s?domain=%s", packagename, domain))

	if err != nil {
		return err
	}

	pkg, err := loadjson(filepath.Join(target.Mapping, ".gsmake.json"))

	if err != nil {
		return err
	}

	return pkg.Properties.Query(name, val)
}

// RootFS get rootfs object
func (runner *Runner) RootFS() vfs.RootFS {
	return runner.rootfs
}

// StartDir get runner start dir
func (runner *Runner) StartDir() string {
	return runner.startdir
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

	runner.I("package name :%s", runner.Name())

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
		stream.WriteString(fmt.Sprintf("\t* %s\n", name))

		for i, child := range task.group {

			scope := strings.ToUpper(child.Scope)

			if scope == "" {
				scope = "ALL"
			}

			stream.WriteString(
				fmt.Sprintf(
					"\t\t%d). package:     %s\n\t\t    description: %s\n \t\t    scope:       %s\n",
					i+1,
					child.Project,
					child.Description,
					scope,
				),
			)
		}

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

	domain := ""
	tokens := strings.SplitN(name, ":", 2)

	if len(tokens) == 2 {

		name = tokens[1]

		domain = tokens[0]
	}

	//DFS Topo sort

	if group, ok := runner.tasks[name]; ok {

		result, err := group.topoShort(runner)

		runner.unmark()

		if err != nil {
			return err
		}

		for _, group := range result {
			if err := group.invoke(runner, domain, args...); err != nil {
				return err
			}
		}

		return nil
	}

	return gserrors.Newf(ErrTask, "unknown task :%s", name)
}
