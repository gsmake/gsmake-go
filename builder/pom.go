package builder

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/gsdocker/gserrors"
)

// ImportPOM .
type ImportPOM struct {
	Name    string // import project Name
	Version string // import project Version
}

// TaskPOM .
type TaskPOM struct {
	Dependency string      `json:"prev"` // projects task name
	Project    *ProjectPOM // project belongs to
}

// ProjectPOM gsmake project
type ProjectPOM struct {
	Name    string              // project universal name
	Version string              // project version
	Import  []*ImportPOM        // project import projects
	Task    map[string]*TaskPOM // project defined task
	Path    string              `json:"-"` // project fullpath
}

func (builder *Builder) loadPOM(path string) (*ProjectPOM, error) {
	configpath := filepath.Join(path, ".gsmake.json")

	content, err := ioutil.ReadFile(configpath)

	if err != nil {
		return nil, gserrors.Newf(err, "load config file err\n\t%s", configpath)
	}

	config := &ProjectPOM{
		Path:    path,
		Version: "current",
	}

	err = json.Unmarshal(content, &config)

	return config, err
}

func (builder *Builder) loadProject(path string) (*ProjectPOM, error) {

	pom, err := builder.loadPOM(path)

	if err != nil {
		return nil, err
	}

	if pom, ok := builder.projects[pom.Name]; ok {
		builder.D("skipp load project %s -- already loaded", pom.Name)
		return pom, nil
	}

	builder.I("loading project %s:%s\n\tpath :%s", pom.Name, pom.Version, path)

	if err := builder.circularLoadingCheck(pom.Name); err != nil {
		return nil, err
	}

	builder.loading = append(builder.loading, pom)

	for _, importPOM := range pom.Import {
		project, err := builder.processImport(importPOM)

		if err != nil {
			return nil, err
		}

		builder.projects[project.Name] = project
	}

	// processing register task

	for name, taskPOM := range pom.Task {
		taskPOM.Project = pom
		builder.tasks[name] = append(builder.tasks[name], taskPOM)
	}

	builder.loading = builder.loading[:len(builder.loading)-1]

	return pom, nil
}

func (builder *Builder) processImport(pom *ImportPOM) (*ProjectPOM, error) {

	if pom.Version == "" {
		pom.Version = "current"
	}

	builder.I("handle import %s:%s", pom.Name, pom.Version)

	path, err := builder.searchProject(pom.Name, pom.Version)

	if err != nil {
		return nil, err
	}

	return builder.loadProject(path)
}
