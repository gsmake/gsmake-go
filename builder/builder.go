package builder

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gsdocker/gsconfig"
	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
	"github.com/gsdocker/gsos"
)

// Builder gsmake builder object
type Builder struct {
	gslogger.Log                        // Mixin gslogger .
	Root         string                 // gsmake Root path
	Name         string                 // build project
	Path         string                 // Root project path
	gopath       string                 // gsmake build path
	binarypath   string                 // builder binary path
	projects     map[string]*ProjectPOM // loaded project collection
	tasks        map[string][]*TaskPOM  // tasks
	loading      []*ProjectPOM          // loading projects
	tpl          *template.Template     // code generate tmplate
}

// NewBuilder create new builder for project
func NewBuilder(root string) (*Builder, error) {

	fullpath, err := filepath.Abs(root)

	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(fullpath, 0755); err != nil {
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

	tpl, err := template.New("golang").Funcs(funcs).Parse(tpl)

	if err != nil {
		return nil, err
	}

	builder := &Builder{
		Log:      gslogger.Get("gsmake"),
		Root:     fullpath,
		projects: make(map[string]*ProjectPOM),
		tasks:    make(map[string][]*TaskPOM),
		tpl:      tpl,
	}

	return builder, nil
}

// Prepare prepare for project indicate by path
func (builder *Builder) Prepare(path string) error {

	fullpath, err := filepath.Abs(path)

	if err != nil {
		return gserrors.Newf(err, "get fullpath -- failed \n\t%s", path)
	}

	builder.gopath = filepath.Join(fullpath, gsconfig.String("gsmake.builder.gopath", ".builder"))
	builder.binarypath = filepath.Join(builder.gopath, "bin", "builder"+gsos.ExeSuffix)

	builder.Path = fullpath

	project, err := builder.loadProject(fullpath)

	if err != nil {
		return err
	}

	builder.Name = project.Name
	builder.projects[project.Name] = project

	return builder.link()

}

// Create create builder project
func (builder *Builder) Create() error {

	builder.I("generate builder src files ...")

	srcRoot := filepath.Join(builder.gopath, "src", "__gsmake")

	if gsos.IsExist(srcRoot) {
		err := os.RemoveAll(srcRoot)

		if err != nil {
			return gserrors.Newf(err, "remove src directory error")
		}
	}

	err := os.MkdirAll(srcRoot, 0755)

	if err != nil {
		return gserrors.Newf(err, "mk src directory error")
	}

	err = builder.genSrcFile(builder, filepath.Join(srcRoot, "main.go"), "main.go")

	if err != nil {
		return err
	}

	i := 0

	for _, project := range builder.projects {

		if len(project.Task) == 0 {
			continue
		}

		err := builder.genSrcFile(project, filepath.Join(srcRoot, fmt.Sprintf("proj_%d.go", i)), "project.go")

		if err != nil {
			return err
		}

		i++
	}

	builder.I("generate builder src files -- success")

	return builder.compileBuilder(srcRoot)
}

// Run run builder
func (builder *Builder) Run(args ...string) error {

	cmd := exec.Command(builder.binarypath, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (builder *Builder) compileBuilder(srcRoot string) error {

	gopath := os.Getenv("GOPATH")

	newgopath := fmt.Sprintf("%s%s%s", builder.gopath, string(os.PathListSeparator), gopath)

	err := os.Setenv("GOPATH", newgopath)

	if err != nil {
		return gserrors.Newf(err, "set new gopath error\n\t%s", builder.gopath)
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

	cmd := exec.Command("go", "build", "-o", builder.binarypath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (builder *Builder) genSrcFile(context interface{}, path string, tplname string) error {

	var buff bytes.Buffer

	builder.I("generate :%s", path)

	if err := builder.tpl.ExecuteTemplate(&buff, tplname, context); err != nil {
		return gserrors.Newf(err, "generate main.go error")
	}

	// var err error
	// bytes, err := format.Source(buff.Bytes())
	//
	// if err != nil {
	// 	return gserrors.Newf(err, "generate src file error\n\tfile:%s", path)
	// }

	err := ioutil.WriteFile(path, buff.Bytes(), 0644)

	if err != nil {
		return gserrors.Newf(err, "generate src file error\n\tfile:%s", path)
	}

	return nil
}

func (builder *Builder) link() error {

	for _, pom := range builder.projects {
		if err := builder.linkProject(pom); err != nil {
			return err
		}
	}

	return nil
}

func (builder *Builder) linkProject(pom *ProjectPOM) error {

	for _, project := range builder.projects {
		if project != pom && strings.HasPrefix(pom.Name, project.Name) {
			if err := builder.linkProject(project); err != nil {
				return err
			}

			break
		}
	}

	linkdir := filepath.Join(builder.gopath, "src", pom.Name)

	if gsos.IsExist(linkdir) {
		if gsos.SameFile(linkdir, pom.Path) {
			builder.I("link project %s:%s -- already exist", pom.Name, pom.Version)
			return nil
		}

		return gserrors.Newf(ErrProject, "duplicate project %s:%s link\n\tone :%s\n\ttwo :%s", pom.Name, pom.Version, linkdir, pom.Path)
	}

	builder.I("link  %s:%s\n\tfrom :%s\n\tto:%s", pom.Name, pom.Version, pom.Path, linkdir)

	err := os.MkdirAll(filepath.Dir(linkdir), 0755)

	if err != nil {
		return err
	}

	err = os.Symlink(pom.Path, linkdir)

	if err != nil {
		return err
	}
	builder.I("link project -- success")

	return nil
}

func (builder *Builder) searchProject(name, version string) (string, error) {

	if version == "" {
		version = "current"
	}

	builder.I("search project %s:%s", name, version)

	// search global repo
	globalpath := filepath.Join(builder.Root, "src", name, version)

	builder.I("search path %s", globalpath)

	if !gsos.IsDir(globalpath) {
		// TODO: invoke download processing
		return "", gserrors.Newf(ErrNotFound, "project %s:%s -- not found", name, version)
	}

	builder.I("search project %s:%s -- found", name, version)

	return globalpath, nil
}

func (builder *Builder) circularLoadingCheck(name string) error {
	var stream bytes.Buffer

	for _, pom := range builder.loading {
		if pom.Name == name || stream.Len() != 0 {
			stream.WriteString(fmt.Sprintf("\t%s import\n", pom.Name))
		}
	}

	if stream.Len() != 0 {
		return gserrors.Newf(ErrProject, "circular package import :\n%s\t%s", stream.String(), name)
	}

	return nil
}
