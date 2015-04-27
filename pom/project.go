package pom

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/gsdocker/gsconfig"
	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gsos"
)

// Task .
type Task struct {
	Func        string // golang function name
	Description string // task description
	Dependency  string // depend prev task
}

// Project .
type Project struct {
	Name       string            // Project universal name
	Version    string            // Project version string
	Task       map[string]*Task  // project declare task map
	Properties map[string]string // properties
}

// NewProject create new project
// @param path project path
func NewProject(path string) (*Project, error) {

	fullpath, err = filepath.Abs(path)

	if err != nil {
		return nil, gserrors.Newf(err, "gsmake project not exist : \n\t%s", path)
	}

	path = fullpath

	if !gsos.IsDir(path) {
		return nil, gserrors.Newf(ErrNotFound, "gsmake project not exist : \n\t%s", path)
	}

	settingfile := filepath.Join(path, gsconfig.String("gsmake.setting.filename", ".gsmake.json"))

	if !gsos.IsExist(settingfile) {
		return nil, gserrors.Newf(ErrNotFound, "gsmake setting file not exist :\n\t%s", settingfile)
	}

	content, err := ioutil.ReadFile(settingfile)

	if err != nil {
		return nil, gserrors.Newf(err, "read project setting file err :\n\tfile: %s\n\terr: %s", settingfile, err)
	}

	project := &Project{}

	err = json.Unmarshal(content, project)

	return project, err
}
