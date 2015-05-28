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

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
	"github.com/gsdocker/gsos"
	"github.com/gsmake/gsmake/vfs"
)

// AOTCompiler aot compiler for package
type AOTCompiler struct {
	gslogger.Log                     // Mixin gslogger .
	binarypath   string              // binary path
	tpl          *template.Template  // code generate tmplate
	rootfs       vfs.RootFS          // vfs
	rootpath     string              // rootpath
	target       string              // package vfs path
	packages     map[string]*Package // loaded packages
}

// Compile .
func Compile(rootpath string, target string) (*AOTCompiler, error) {

	loader, err := load(rootpath, target)

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
		tpl:      tpl,
		rootfs:   loader.rootfs,
		target:   loader.targetpath,
		rootpath: rootpath,
		packages: loader.packages,
	}

	compiler.binarypath = filepath.Join(compiler.rootfs.TempDir("task"), "runner"+gsos.ExeSuffix)

	return compiler, compiler.compile()
}

func (compiler *AOTCompiler) compile() error {

	srcRoot := filepath.Join(compiler.rootfs.DomainDir("task"), "src", "runner")

	compiler.D("srcroot :%s", srcRoot)

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

	var context = struct {
		RootPath   string
		TargetPath string
	}{
		compiler.rootpath,
		compiler.target,
	}

	err = compiler.gencodes(context, filepath.Join(srcRoot, "main.go"), "main.go")

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

	newgopath := compiler.rootfs.DomainDir("task")

	err := os.Setenv("GOPATH", newgopath)

	if err != nil {
		return gserrors.Newf(err, "set new gopath error\n\t%s", newgopath)
	}

	compiler.D("GOPATH:\n/%s", newgopath)

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

// Run run compiler generate program
func (compiler *AOTCompiler) Run(startdir string, args ...string) error {

	cmd := exec.Command(compiler.binarypath, args...)

	cmd.Dir = startdir

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
