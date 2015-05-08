package gsmake

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	Stage   string // runtimes import flag, default is AOT import
}

// Task package defined task description
type Task struct {
	Prev        string // depend task name
	Description string // task description
	Package     string `json:"-"` // package name which defined this task
}

// Site package host site
type Site struct {
	Name    string // the site's name
	SCM     string // the scm url pattern
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
}

// Loader package loader
type Loader struct {
	gslogger.Log                     // mixin Log APIs
	settings     Settings            // gsmake settings
	packages     map[string]*Package // loaded packages
	checkerOfDCG []*Package          // DCG check stack
	stage        stageType           // loader exec stage
	repository   *Repository         // gsmake repository
}

// Load load package
func Load(homepath string, name string, stage stageType) (*Loader, error) {

	loader := &Loader{
		Log:      gslogger.Get("gsmake"),
		packages: make(map[string]*Package),
		stage:    stage,
	}

	loader.settings.name = name
	loader.settings.setHome(homepath)

	// load respository
	repo, err := loadRepository(loader.settings)

	if err != nil {
		return nil, err
	}

	loader.repository = repo

	err = loader.load()

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (loader *Loader) prepare() error {

	// for _, pkg := range loader.packages {
	// 	if err := loader.linkpackage(pkg); err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func (loader *Loader) load() error {
	path := loader.settings.workpath(loader.settings.name)

	pkg, err := loader.loadpackage(loader.settings.name, path)

	if err != nil {
		return err
	}

	loader.packages[pkg.Name] = pkg

	return nil
}

func (loader *Loader) loadpackage(name string, path string) (*Package, error) {
	// check if already loaded this package

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
		}, nil
	}

	pkg, err := loader.loadjson(jsonfile)

	if err != nil {
		return nil, err
	}

	if name != "" && pkg.Name != name {
		return nil, gserrors.Newf(ErrPackage, "package name must be %s\n\tpath :%s", name, path)
	}

	loader.checkerOfDCG = append(loader.checkerOfDCG, pkg)

	for _, importir := range pkg.Import {

		if importir.Version == "" {
			importir.Version = "current"
		}

		stage, err := stageParse(importir.Stage)

		if err != nil {
			return nil, gserrors.Newf(err, "parse import [%s] scope error\n\t%s", importir.Name, path)
		}

		if (stage & loader.stage) == 0 {
			continue
		}

		importpath, err := loader.repository.Search(importir.Name, importir.Version)

		if err != nil {
			return nil, err
		}

		importpkg, err := loader.loadpackage(importir.Name, importpath)

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
