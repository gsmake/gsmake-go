package gsmake

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
	"github.com/gsdocker/gsos"
	"github.com/gsmake/gsmake/fs"
	"github.com/gsmake/gsmake/vfs"
)

// gsmake predeined environment variables
const (
	EnvHome = "GSMAKE_HOME"
)

// Errors .
var (
	ErrLoad = errors.New("load package error")
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
	Name    string // import package name
	Version string // import package version
	Domain  string `json:"scope"` // runtimes import flag, default is AOT import
	SCM     string // the source control manager type
	URL     string // remote url
}

// Task package defined task description
type Task struct {
	Prev        string // depend task name
	Description string // task description
	Package     string `json:"-"` // package name which defined this task
}

// Package describe a gsmake package object
type Package struct {
	Name       string          // package name string
	Import     []Import        // package import field
	Task       map[string]Task // package defined task
	Properties Properties      // properties
	vfspath    string          // vfs path
}

// Loader package loader
type Loader struct {
	gslogger.Log                     // mixin Log APIs
	packages     map[string]*Package // loaded packages
	checkerOfDCG []*Package          // DCG check stack
	rootfs       vfs.RootFS          // vfs object
	targetpath   string              // the loading package path
}

func load(rootpath string, target string) (*Loader, error) {

	fullpath, err := filepath.Abs(target)

	if err != nil {
		return nil, err
	}

	rootfs, err := vfs.New(rootpath, fullpath)

	if err != nil {
		return nil, err
	}

	loader := &Loader{
		Log:        gslogger.Get("loader"),
		targetpath: fullpath,
		packages:   make(map[string]*Package),
		rootfs:     rootfs,
	}

	loader.I("load package ...")

	start := time.Now()

	err = loader.load()

	loader.I("load package -- success %s", time.Now().Sub(start))

	if err != nil {
		return nil, err
	}

	return loader, nil
}

func (loader *Loader) load() error {

	jsonfile := filepath.Join(loader.targetpath, ".gsmake.json")

	if !fs.Exists(jsonfile) {

		return gserrors.Newf(
			ErrLoad,
			"target package not exists or is not a gsmake package\n\t%s",
			loader.targetpath,
		)
	}

	pkg, err := loadjson(jsonfile)

	if err != nil {
		return err
	}

	pkg, err = loader.loadpackagev2("", pkg.Name, loader.targetpath)

	if err != nil {
		return err
	}

	target := fmt.Sprintf("gsmake://%s?domain=task", pkg.Name)

	src := fmt.Sprintf("file://%s?version=current", loader.targetpath)

	pkg.vfspath = target

	loader.packages[pkg.vfspath] = pkg

	if !loader.rootfs.Mounted(src, target) {
		if err := loader.rootfs.Mount(src, target); err != nil {
			return err
		}
	}

	// try load gsmake
	if _, ok := loader.packages["gsmake://github.com/gsmake/gsmake?domain=task"]; !ok {

		pkg, err := loader.loadpackage(Import{
			Name:    "github.com/gsmake/gsmake",
			Version: "v2.0",
			SCM:     "git",
			Domain:  "task",
		})

		if err != nil {
			return gserrors.Newf(err, "load package github.com/gsmake/gsmake error")
		}

		loader.packages[pkg.vfspath] = pkg
	}

	// dismount not loaded packages

	entries, err := loader.rootfs.List()

	if err != nil {
		return gserrors.Newf(err, "list vfs nodes error")
	}

	for k := range loader.packages {
		loader.D("loaded package :%s", k)
	}

	for target := range entries {
		if _, ok := loader.packages[target]; !ok {
			loader.I("dismount :%s", target)
			loader.rootfs.Dismount(target)
		}
	}

	return nil
}

func (loader *Loader) checkDCG(name string) error {

	var stream bytes.Buffer

	for _, pkg := range loader.checkerOfDCG {
		if pkg.Name == name || stream.Len() != 0 {
			stream.WriteString(fmt.Sprintf("\t%s import\n", pkg.Name))
		}
	}

	if stream.Len() != 0 {
		return gserrors.Newf(ErrLoad, "circular package import :\n%s\t%s", stream.String(), name)
	}

	return nil
}

func (loader *Loader) loadpackage(i Import) (*Package, error) {

	target := fmt.Sprintf("gsmake://%s?domain=%s", i.Name, i.Domain)

	src := fmt.Sprintf("%s://%s?version=%s", i.SCM, i.Name, i.Version)

	if !loader.rootfs.Mounted(src, target) {
		if err := loader.rootfs.Mount(src, target); err != nil {
			return nil, err
		}
	}

	_, entry, err := loader.rootfs.Open(target)

	if err != nil {
		return nil, err
	}

	if pkg, ok := loader.packages[target]; ok {
		return pkg, nil
	}

	// DCG check
	if err := loader.checkDCG(target); err != nil {
		return nil, err
	}

	importpkg, err := loader.loadpackagev2(i.Domain, i.Name, entry.Mapping)

	importpkg.vfspath = target

	return importpkg, err
}

func (loader *Loader) loadpackagev2(domain, name, fullpath string) (*Package, error) {

	jsonfile := filepath.Join(fullpath, ".gsmake.json")

	if !gsos.IsExist(jsonfile) {
		// this package is a traditional golang package
		return &Package{
			Name: name,
		}, nil
	}

	pkg, err := loadjson(jsonfile)

	if err != nil {
		return nil, err
	}

	loader.checkerOfDCG = append(loader.checkerOfDCG, pkg)

	for _, v := range pkg.Import {

		if _, ok := loader.packages[v.Name]; ok {
			continue
		}

		if v.Version == "" {
			v.Version = "current"
		}

		if v.SCM == "" {

			u, err := url.Parse(fmt.Sprintf("https://%s", v.Name))

			if err != nil {
				return nil, gserrors.Newf(err, "%s invalid import package :%s", name, v.Name)
			}

			v.SCM = loader.rootfs.Protocol(u.Host)
		}

		domains := strings.Split(v.Domain, "|")

		if len(domains) == 1 && domains[0] == "" {
			domains = []string{"task"}
		}

		for _, d := range domains {

			if domain != d && domain != "" {

				continue
			}

			v.Domain = d

			importpkg, err := loader.loadpackage(v)

			if err != nil {
				return nil, err
			}

			loader.packages[importpkg.vfspath] = importpkg
		}
	}

	loader.checkerOfDCG = loader.checkerOfDCG[:len(loader.checkerOfDCG)-1]

	return pkg, nil
}

func loadjson(file string) (*Package, error) {

	content, err := ioutil.ReadFile(file)

	if err != nil {
		return nil, gserrors.Newf(err, "load config file err\n\t%s", file)
	}

	var config *Package

	err = json.Unmarshal(content, &config)

	if err != nil {
		return nil, gserrors.Newf(err, "unmarshal .gsmake.json file error\n\tfile:%s", file)
	}

	return config, nil
}
