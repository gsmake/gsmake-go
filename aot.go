package gsmake

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"go/format"

	"github.com/gsdocker/gsconfig"
	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
	"github.com/gsdocker/gsos"
)

// AOTCompiler aot compiler for package
type AOTCompiler struct {
	gslogger.Log                     // Mixin gslogger .
	root         string              // gsmake Root path
	name         string              // build project
	path         string              // compile package path
	gopath       string              // compile gopath
	binarypath   string              // generate task program path
	tpl          *template.Template  // code generate tmplate
	packages     map[string]*Package // load packages
}

// Compile invoke aot compile for current package which path is ${packagedir}
func Compile(root string, packagedir string) (*AOTCompiler, error) {

	loader, err := Load(root, packagedir, scopeTask)

	if err != nil {
		return nil, err
	}

	funcs := template.FuncMap{
		"taskname": func(name string) string {
			return "Task" + strings.Title(name)
		},
		"ospath": func(name string) string {
			return strings.Replace(name, "\\", "\\\\", -1)
		},
	}

	tpl, err := template.New("golang").Funcs(funcs).Parse(codegen)

	if err != nil {
		return nil, err
	}

	compiler := &AOTCompiler{
		Log:      gslogger.Get("gsmake"),
		root:     loader.root,
		name:     loader.pkg.Name,
		path:     loader.path,
		gopath:   filepath.Join(loader.path, gsconfig.String("gsmake.taskdir", ".task")),
		tpl:      tpl,
		packages: loader.packages,
	}

	compiler.binarypath = filepath.Join(compiler.gopath, "bin", "__gsmake_task"+gsos.ExeSuffix)

	return compiler, compiler.compile()
}

// Run run compiler generate program
func (compiler *AOTCompiler) Run(args ...string) error {

	gopath := os.Getenv("GOPATH")

	err := os.Setenv("GOPATH", filepath.Join(compiler.path, gsconfig.String("gsmake.rundir", ".run")))

	if err != nil {
		return gserrors.Newf(err, "set new gopath error\n\t%s", compiler.gopath)
	}

	defer func() {
		os.Setenv("GOPATH", gopath)
	}()

	cmd := exec.Command(compiler.binarypath, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func (compiler *AOTCompiler) compile() error {

	srcRoot := filepath.Join(compiler.path, gsconfig.String("gsmake.taskdir", ".task"), "src/gsmake.task")

	if gsos.IsExist(srcRoot) {
		err := os.RemoveAll(srcRoot)

		if err != nil {
			return gserrors.Newf(err, "remove gsmake.task dir error")
		}
	}

	err := os.MkdirAll(srcRoot, 0755)

	if err != nil {
		return gserrors.Newf(err, "mk src directory error")
	}

	var context struct {
		Name string
		Path string
		Root string
	}

	context.Name = compiler.name
	context.Path = compiler.path
	context.Root = compiler.root

	err = compiler.gencodes(&context, filepath.Join(srcRoot, "main.go"), "main.go")

	if err != nil {
		return err
	}

	i := 0

	for _, pkg := range compiler.packages {

		if len(pkg.Task) == 0 {
			continue
		}

		err := compiler.gencodes(pkg, filepath.Join(srcRoot, fmt.Sprintf("proj_%d.go", i)), "project.go")

		if err != nil {
			return err
		}

		i++
	}

	err = compiler.genbinary(srcRoot)

	if err != nil {
		return gserrors.Newf(err, "generate binary error")
	}

	return nil
}

func (compiler *AOTCompiler) genbinary(srcRoot string) error {

	gopath := os.Getenv("GOPATH")

	newgopath := compiler.gopath //fmt.Sprintf("%s%s%s", compiler.gopath, string(os.PathListSeparator), gopath)

	err := os.Setenv("GOPATH", newgopath)

	if err != nil {
		return gserrors.Newf(err, "set new gopath error\n\t%s", compiler.gopath)
	}

	defer func() {
		os.Setenv("GOPATH", gopath)
	}()

	currentDir, err := filepath.Abs("./")

	if err != nil {
		return gserrors.Newf(err, "get current dir error")
	}

	err = os.Chdir(srcRoot)

	if err != nil {
		return gserrors.Newf(err, "change current dir error\n\tto:%s", srcRoot)
	}

	defer func() {
		os.Chdir(currentDir)
	}()

	cmd := exec.Command("go", "build", "-o", compiler.binarypath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (compiler *AOTCompiler) gencodes(context interface{}, path string, tplname string) error {

	var buff bytes.Buffer

	if err := compiler.tpl.ExecuteTemplate(&buff, tplname, context); err != nil {
		return gserrors.Newf(err, "generate main.go error")
	}

	// var err error
	bytes, err := format.Source(buff.Bytes())

	if err != nil {
		return gserrors.Newf(err, "generate src file error\n\tfile:%s", path)
	}

	err = ioutil.WriteFile(path, bytes, 0644)

	if err != nil {
		return gserrors.Newf(err, "generate src file error\n\tfile:%s", path)
	}

	return nil
}
