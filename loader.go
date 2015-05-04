package gsmake

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
	"github.com/gsdocker/gsos"
)

// errors
var (
	ErrNotFound = errors.New("project not found")
	ErrProject  = errors.New("project setting error")
)

// Loader project loader
type Loader struct {
	gslogger.Log                        // Mixin gslogger APIs
	root         string                 // gsmake root path
	path         string                 // project root path
	buildpath    string                 // build path
	projects     map[string]*ProjectPOM // loaded project collection
	loading      []*ProjectPOM          // loading projects
	runtimes     bool                   // runtime loader flag
	downloader   *Downloader            // project downloader
}

// ImportPOM .
type ImportPOM struct {
	Name     string // import project Name
	Version  string // import project Version
	Runtimes bool   // runtime import flag
}

// TaskPOM .
type TaskPOM struct {
	Prev    string      // projects task name
	Project *ProjectPOM // project belongs to
}

// ProjectPOM gsmake project
type ProjectPOM struct {
	Name    string              // project universal name
	Version string              // project version
	Import  []*ImportPOM        // project import projects
	Task    map[string]*TaskPOM // project defined task
	Path    string              `json:"-"` // project fullpath
}

// NewLoader create new project loadr
func NewLoader(root string, runtimes bool) (*Loader, error) {
	path, err := filepath.Abs(root)

	if err != nil {
		return nil, err
	}

	return &Loader{
		Log:        gslogger.Get("gsmake"),
		projects:   make(map[string]*ProjectPOM),
		root:       path,
		runtimes:   runtimes,
		downloader: NewDownloader(),
	}, nil
}

// Projects .
func (loader *Loader) Projects() map[string]*ProjectPOM {
	return loader.projects
}

// Load load project
func (loader *Loader) Load(path string, buildpath string) (*ProjectPOM, error) {
	fullpath, err := filepath.Abs(path)

	if err != nil {
		return nil, err
	}

	fullbuildpath, err := filepath.Abs(buildpath)

	if err != nil {
		return nil, err
	}

	loader.buildpath = fullbuildpath

	project, err := loader.loadProject(fullpath)

	if err != nil {
		return nil, err
	}

	err = loader.link()

	if err != nil {
		return nil, err
	}

	return project, nil
}

func (loader *Loader) circularLoadingCheck(name string) error {
	var stream bytes.Buffer

	for _, pom := range loader.loading {
		if pom.Name == name || stream.Len() != 0 {
			stream.WriteString(fmt.Sprintf("\t%s import\n", pom.Name))
		}
	}

	if stream.Len() != 0 {
		return gserrors.Newf(ErrProject, "circular package import :\n%s\t%s", stream.String(), name)
	}

	return nil
}

func (loader *Loader) loadProject(path string) (*ProjectPOM, error) {

	pom, err := loader.loadPOM(path)

	if err != nil {
		return nil, err
	}

	if pom, ok := loader.projects[pom.Name]; ok {
		loader.D("skipp load project %s -- already loaded", pom.Name)
		return pom, nil
	}

	loader.I("loading project %s:%s\n\tpath :%s", pom.Name, pom.Version, path)

	if err := loader.circularLoadingCheck(pom.Name); err != nil {
		return nil, err
	}

	loader.loading = append(loader.loading, pom)

	for _, importPOM := range pom.Import {

		if importPOM.Runtimes != loader.runtimes {
			continue
		}

		project, err := loader.processImport(importPOM)

		if err != nil {
			return nil, err
		}

		loader.projects[project.Name] = project
	}

	for _, taskPOM := range pom.Task {
		taskPOM.Project = pom
	}

	loader.loading = loader.loading[:len(loader.loading)-1]

	return pom, nil
}

func (loader *Loader) processImport(pom *ImportPOM) (*ProjectPOM, error) {

	if pom.Version == "" {
		pom.Version = "current"
	}

	loader.I("handle import %s:%s", pom.Name, pom.Version)

	path, err := loader.searchProject(pom.Name, pom.Version)

	if err != nil {
		return nil, err
	}

	return loader.loadProject(path)
}

func (loader *Loader) loadPOM(path string) (*ProjectPOM, error) {
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

func (loader *Loader) searchProject(name, version string) (string, error) {

	if version == "" {
		version = "current"
	}

	loader.I("search project %s:%s", name, version)

	// search global repo
	globalpath := filepath.Join(loader.root, "src", name, version)

	loader.I("search path %s", globalpath)

	if !gsos.IsDir(globalpath) {
		err := loader.downloader.Download(name, version, globalpath)

		if err != nil {
			return "", gserrors.Newf(err, "project %s:%s -- not found", name, version)
		}
	}

	loader.I("search project %s:%s -- found", name, version)

	return globalpath, nil
}

func (loader *Loader) link() error {

	for _, pom := range loader.projects {
		if err := loader.linkProject(pom); err != nil {
			return err
		}
	}

	return nil
}

func (loader *Loader) linkProject(pom *ProjectPOM) error {

	for _, project := range loader.projects {
		if project != pom && strings.HasPrefix(pom.Name, project.Name) {
			if err := loader.linkProject(project); err != nil {
				return err
			}

			break
		}
	}

	linkdir := filepath.Join(loader.buildpath, "src", pom.Name)

	if gsos.IsExist(linkdir) {
		if gsos.SameFile(linkdir, pom.Path) {
			loader.I("link project %s:%s -- already exist", pom.Name, pom.Version)
			return nil
		}

		return gserrors.Newf(ErrProject, "duplicate project %s:%s link\n\tone :%s\n\ttwo :%s", pom.Name, pom.Version, linkdir, pom.Path)
	}

	loader.I("link  %s:%s\n\tfrom :%s\n\tto:%s", pom.Name, pom.Version, pom.Path, linkdir)

	err := os.MkdirAll(filepath.Dir(linkdir), 0755)

	if err != nil {
		return err
	}

	err = os.Symlink(pom.Path, linkdir)

	if err != nil {
		return err
	}
	loader.I("link project -- success")

	return nil
}
