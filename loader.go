package gsmake

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
	"github.com/gsdocker/gsos"
)

type stageType int

const (
	stageTask stageType = (1 << iota)
	stageRuntimes

	scopAll = stageTask | stageRuntimes
)

func stageParse(text string) (stageType, error) {

	var result stageType

	for _, v := range strings.Split(text, "|") {
		switch v {
		case "task":
			result |= stageTask
		case "runtimes":
			result |= stageRuntimes
		case "":
			continue
		default:
			return result, gserrors.Newf(ErrPackage, "unknown stage type :%s", v)
		}
	}

	if result == 0 {
		result = stageTask
	}

	return result, nil
}

// Properties .
type Properties map[string]interface{}

// Expand rewrites content to replace ${k} with properties[k] for each key k in match.
func Expand(content string, properties Properties) string {
	for k, v := range properties {

		if stringer, ok := v.(fmt.Stringer); ok {
			fmt.Println(stringer.String())
			content = strings.Replace(content, "${"+k+"}", stringer.String(), -1)
		} else {
			content = strings.Replace(content, "${"+k+"}", fmt.Sprintf("%v", v), -1)
		}

	}
	return content
}

// Import the gsmake import instruction description
type Import struct {
	Name    string // import package name
	Version string // import package version
	Stage   string `json:"scope"` // runtimes import flag, default is AOT import
}

// Task package defined task description
type Task struct {
	Prev        string // depend task name
	Description string // task description
	Package     string `json:"-"` // package name which defined this task
}

// Site package host site
type Site struct {
	SCM     string // the site's name
	URL     string // the scm url pattern
	Package string // the site support package name pattern
}

// Package describe a gsmake package object
type Package struct {
	Name       string          // package name string
	Version    string          // package version string
	Import     []Import        // package import field
	Task       map[string]Task // package defined task
	Site       map[string]Site // vcs site list
	Properties Properties      // properties
	origin     string          // package origin path
}

// Loader package loader
type Loader struct {
	gslogger.Log                          // mixin Log APIs
	packages     map[string]*Package      // loaded packages
	checkerOfDCG []*Package               // DCG check stack
	stage        stageType                // loader exec stage
	repository   *Repository              // gsmake repository
	name         string                   // load package name
	homepath     string                   // gsmake home path
	rootpackage  *Package                 // load root pacakge
	nocached     bool                     // load cache flag
	cached       map[string]string        // cached package's index
	importdir    func(name string) string // calculator of importdir
	indexfile    string                   // indexfile
}

// Load load package
func Load(homepath string, path string, stage stageType, nocached bool, imports []Import) (*Loader, error) {

	loader := &Loader{
		Log:      gslogger.Get("gsmake"),
		packages: make(map[string]*Package),
		stage:    stage,
		nocached: nocached,
	}

	if stage == stageTask {
		loader.importdir = func(name string) string {
			return TaskStageImportDir(loader.homepath, loader.name, name)
		}
	} else {
		loader.importdir = func(name string) string {
			return RuntimesStageImportDir(loader.homepath, loader.name, name)
		}
	}

	var err error

	loader.homepath, err = filepath.Abs(homepath)

	if err != nil {
		return nil, err
	}

	// load respository
	repo, err := openRepository(loader.homepath)

	if err != nil {
		return nil, err
	}

	loader.repository = repo

	err = loader.load(path)

	if err != nil {
		return nil, err
	}

	for _, v := range imports {
		pkg, err := loader.loadpackage(v.Name, v.Version)

		if err != nil {
			return nil, err
		}

		loader.packages[pkg.Name] = pkg
	}

	return loader, nil
}

// LoadPackage .
func (loader *Loader) LoadPackage(name string, version string) (*Package, error) {

	pkg, err := loader.loadpackage(name, version)

	if err != nil {
		return nil, err
	}

	loader.packages[pkg.Name] = pkg

	loader.savecache()

	return pkg, nil
}

func (loader *Loader) loadpackage(name string, version string) (*Package, error) {
	cachedpath := loader.importdir(name)

	cachedversion, ok := loader.cached[name]

	if loader.nocached || !ok || cachedversion != version {

		err := loader.repository.Copy(name, version, cachedpath)

		if err != nil {
			return nil, err
		}
	}

	// update cached index
	loader.cached[name] = version

	importpkg, err := loader.loadpackagev2(version, cachedpath)

	return importpkg, err
}

func (loader *Loader) savecache() {
	// marshal cache index
	content, err := json.Marshal(loader.cached)

	if err != nil {
		loader.W("save cache index error\n%s", err)
		return
	}

	if err := ioutil.WriteFile(loader.indexfile, content, 0644); err != nil {
		loader.W("save cache index error\n%s", err)
	}
}

func (loader *Loader) loadcache() error {
	var indexfile string

	if loader.stage == stageTask {
		indexfile = TaskStageImportDir(loader.homepath, loader.name, "")
	} else {
		indexfile = RuntimesStageImportDir(loader.homepath, loader.name, "")
	}

	indexfile = filepath.Join(indexfile, ".cached")

	loader.indexfile = indexfile

	if gsos.IsExist(indexfile) {
		// load cached package'a index
		content, err := ioutil.ReadFile(indexfile)

		if err != nil {
			return gserrors.Newf(err, "read cache index file error")
		}

		if err := json.Unmarshal(content, &loader.cached); err != nil {
			return gserrors.Newf(err, "read cache index file error")
		}
	} else {
		loader.cached = make(map[string]string)
	}

	return nil
}

func (loader *Loader) load(path string) error {

	var err error

	path, err = filepath.Abs(path)

	if err != nil {
		return gserrors.Newf(err, "get load package fullpath error :%s", path)
	}

	// first of all read the package's name
	jsonfile := filepath.Join(path, NameConfigFile)

	pkg, err := loader.loadjson(jsonfile)

	if err != nil {
		return err
	}

	loader.name = pkg.Name

	if err := loader.loadcache(); err != nil {
		return err
	}

	// try link current package to workspace
	targetdir := WorkspaceImportDir(loader.homepath, loader.name)

	if !gsos.SameFile(path, targetdir) {

		// if target is exist
		if gsos.IsExist(targetdir) {
			gsos.RemoveAll(targetdir)
		} else {
			// try make parent directory
			err = os.MkdirAll(filepath.Dir(targetdir), 0755)

			if err != nil {
				return gserrors.Newf(err, "create workspace error")
			}
		}

		// do symbol link
		err = os.Symlink(path, targetdir)

		if err != nil {
			return gserrors.Newf(err, "link package to workspace error")
		}
	}

	// now try doload package

	pkg, err = loader.loadpackagev2(loader.name, path)

	if err != nil {
		return err
	}

	loader.packages[pkg.Name] = pkg

	// check if had loaded package github.com/gsdocker/gsmake

	if _, ok := loader.packages["github.com/gsdocker/gsmake"]; !ok {

		pkg, err := loader.loadpackage("github.com/gsdocker/gsmake", "current")

		if err != nil {
			return gserrors.Newf(err, "load package github.com/gsdocker/gsmake error")
		}

		loader.packages[pkg.Name] = pkg
	}

	loader.savecache()

	return nil
}

func (loader *Loader) loadpackagev2(name string, path string) (*Package, error) {

	if pkg, ok := loader.packages[name]; ok {
		return pkg, nil
	}

	// DCG check
	if err := loader.checkDCG(name); err != nil {
		return nil, err
	}

	jsonfile := filepath.Join(path, NameConfigFile)

	if !gsos.IsExist(jsonfile) {
		// this package is a traditional golang package
		return &Package{
			Name:    name,
			Version: "current",
			origin:  path,
		}, nil
	}

	pkg, err := loader.loadjson(jsonfile)

	if err != nil {
		return nil, err
	}

	pkg.origin = path

	loader.checkerOfDCG = append(loader.checkerOfDCG, pkg)

	for _, importIR := range pkg.Import {

		if _, ok := loader.packages[importIR.Name]; ok {
			continue
		}

		if importIR.Version == "" {
			importIR.Version = "current"
		}

		stage, err := stageParse(importIR.Stage)

		if err != nil {
			return nil, gserrors.Newf(err, "parse import [%s] scope error\n\t%s", importIR.Name, path)
		}

		if (stage & loader.stage) == 0 {
			continue
		}

		importpkg, err := loader.loadpackage(importIR.Name, importIR.Version)

		if err != nil {
			return nil, err
		}

		loader.packages[importpkg.Name] = importpkg
	}

	loader.checkerOfDCG = loader.checkerOfDCG[:len(loader.checkerOfDCG)-1]

	return pkg, nil
}

// Cycline check
func (loader *Loader) checkDCG(name string) error {

	var stream bytes.Buffer

	for _, pkg := range loader.checkerOfDCG {
		if pkg.Name == name || stream.Len() != 0 {
			stream.WriteString(fmt.Sprintf("\t%s import\n", pkg.Name))
		}
	}

	if stream.Len() != 0 {
		return gserrors.Newf(ErrPackage, "circular package import :\n%s\t%s", stream.String(), name)
	}

	return nil
}

func (loader *Loader) loadjson(file string) (*Package, error) {
	content, err := ioutil.ReadFile(file)

	if err != nil {
		return nil, gserrors.Newf(err, "load config file err\n\t%s", file)
	}

	config := &Package{
		Version: "current",
	}

	err = json.Unmarshal(content, &config)

	if err != nil {
		return nil, gserrors.Newf(err, "unmarshal .gsmake.json file error\n\tfile:%s", file)
	}

	return config, nil
}
