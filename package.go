package gsmake

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gsdocker/gsconfig"
	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
	"github.com/gsdocker/gsos"
)

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
	Name     string // import package name
	Version  string // import package version
	Runtimes bool   // runtimes import flag, default is AOT import
}

// Task package defined task description
type Task struct {
	Prev    string // depend task name
	Package string `json:"-"` // package name which defined this task
}

// Package describe a gsmake package object
type Package struct {
	Name        string          // package name string
	Version     string          // package version string
	Import      []Import        // package import field
	Task        map[string]Task // package defined task
	Properties  Properties      // properties
	Path        string          `json:"-"` // package origin path
	Linked      string          `json:"-"` // package linked path
	Traditional bool            `json:"-"` // traditional golang package flag
}

// Loader gsmake package loader
type Loader struct {
	gslogger.Log                     // mixin Log APIs
	pkg          *Package            // root package
	packages     map[string]*Package // loaded packages
	checkerOfDCG []*Package          // DCG check stack
	runtimes     bool                // runtimes loader flag
	root         string              // gsmake root path
	path         string              // load package dir
	properties   Properties          // properties
	downloader   *Downloader         // downloader
}

// Load load package
func Load(root string, packagedir string, runtimes bool) (*Loader, error) {

	rootpath, err := filepath.Abs(root)

	if err != nil {
		return nil, gserrors.Newf(err, "calc gsmake root fullpath error")
	}

	if !gsos.IsExist(rootpath) {
		if err := os.MkdirAll(rootpath, 0755); err != nil {
			return nil, err
		}
	}

	packagepath, err := filepath.Abs(packagedir)

	if err != nil {
		return nil, gserrors.Newf(err, "calc package fullpath error")
	}

	loader := &Loader{
		Log:        gslogger.Get("gsmake"),
		packages:   make(map[string]*Package),
		runtimes:   runtimes,
		root:       rootpath,
		path:       packagepath,
		properties: make(Properties),
		downloader: NewDownloader(),
	}

	err = loader.load()

	if err != nil {

		return nil, err
	}

	return loader, loader.link()
}

// Properties .
func (loader *Loader) Properties() Properties {
	return loader.properties
}

func (loader *Loader) link() error {

	for _, pkg := range loader.packages {
		if err := loader.linkpackage(pkg); err != nil {
			return err
		}
	}

	return nil
}

func (loader *Loader) linkpackage(pkg *Package) error {

	for _, parent := range loader.packages {
		if parent != pkg && strings.HasPrefix(pkg.Name, parent.Name) {
			if err := loader.linkpackage(parent); err != nil {
				return err
			}

			break
		}
	}

	var linkdir string

	if loader.runtimes {
		linkdir = filepath.Join(loader.path, gsconfig.String("gsmake.rundir", ".run"), "src", pkg.Name)
	} else {
		linkdir = filepath.Join(loader.path, gsconfig.String("gsmake.taskdir", ".task"), "src", pkg.Name)
	}

	if gsos.IsExist(linkdir) {
		if gsos.SameFile(linkdir, pkg.Path) {
			loader.D("link project %s:%s -- already exist", pkg.Name, pkg.Version)
			return nil
		}

		return gserrors.Newf(ErrPackage, "duplicate project %s:%s link\n\tone :%s\n\ttwo :%s", pkg.Name, pkg.Version, linkdir, pkg.Path)
	}

	loader.D("link  %s:%s\n\tfrom :%s\n\tto:%s", pkg.Name, pkg.Version, pkg.Path, linkdir)

	err := os.MkdirAll(filepath.Dir(linkdir), 0755)

	if err != nil {
		return err
	}

	err = os.Symlink(pkg.Path, linkdir)

	if err != nil {
		return err
	}

	loader.D("link project -- success")

	return nil
}

func (loader *Loader) load() error {
	pkg, err := loader.loadpackage("", loader.path)

	if err != nil {
		return err
	}

	loader.packages[pkg.Name] = pkg
	loader.pkg = pkg

	return nil
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

func (loader *Loader) loadpackage(name string, fullpath string) (*Package, error) {

	// check if already loaded this package

	if pkg, ok := loader.packages[name]; ok {
		return pkg, nil
	}

	// DCG check
	if err := loader.checkDCG(name); err != nil {
		return nil, err
	}

	jsonfile := filepath.Join(fullpath, gsconfig.String("gsmake.filename", ".gsmake.json"))

	if !gsos.IsExist(jsonfile) {

		// this package is a traditional golang package

		return &Package{
			Name:        name,
			Path:        fullpath,
			Version:     "current",
			Traditional: true,
		}, nil
	}

	pkg, err := loader.loadjson(jsonfile)

	if err != nil {
		return nil, err
	}

	if name != "" && pkg.Name != name {
		return nil, gserrors.Newf(ErrPackage, "package name must be %s\n\tpath :%s", name, fullpath)
	}

	loader.checkerOfDCG = append(loader.checkerOfDCG, pkg)

	for _, importir := range pkg.Import {

		if importir.Version == "" {
			importir.Version = "current"
		}

		if importir.Runtimes != loader.runtimes {
			continue
		}

		importpath, err := loader.searchpackage(importir.Name, importir.Version)

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

	if pkg.Properties != nil {
		for k, v := range pkg.Properties {
			loader.properties[k] = v
		}
	}

	return pkg, nil
}

func (loader *Loader) searchpackage(name string, version string) (string, error) {

	// first search local repo
	repopath := filepath.Join(loader.root, "packages", name, version)

	if gsos.IsExist(repopath) {
		return repopath, nil
	}

	err := loader.downloader.Download(name, version, repopath)

	if err != nil {
		return "", gserrors.Newf(err, "unknown package %s:%s", name, version)
	}

	return repopath, nil
}

func (loader *Loader) loadjson(file string) (*Package, error) {
	content, err := ioutil.ReadFile(file)

	if err != nil {
		return nil, gserrors.Newf(err, "load config file err\n\t%s", file)
	}

	config := &Package{
		Path:    filepath.Dir(file),
		Version: "current",
	}

	err = json.Unmarshal(content, &config)

	return config, err
}
