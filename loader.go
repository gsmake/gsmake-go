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

	"github.com/gsdocker/gsos/fs"
	"github.com/gsmake/gsmake/vfs"
)

// gsmake predeined environment variables
const (
	EnvHome          = "GSMAKE_HOME"
	VersionGSMake    = "release/v2.0"
	PacakgeAnonymous = "github.com/gsmake/gsmake.anonymous"
	Logfmt           = "[$tag] $content"
	LogTimefmt       = ""
)

// Errors .
var (
	ErrLoad = errors.New("load package error")
)

// Loader package loader
type Loader struct {
	gslogger.Log                                // mixin Log APIs
	packages     map[string]map[string]*Package // loaded packages
	checkerOfDCG []*Package                     // DCG check stack
	rootfs       vfs.RootFS                     // vfs object
	targetpath   string                         // the loading package path
	imports      []Import                       // extra imports
}

func load(rootfs vfs.RootFS, imports []Import) (*Loader, error) {

	loader := &Loader{
		Log:        gslogger.Get("loader"),
		targetpath: rootfs.TargetPath(),
		packages:   make(map[string]map[string]*Package),
		rootfs:     rootfs,
		imports:    imports,
	}

	loader.I("load package ...")

	start := time.Now()

	err := loader.load()

	if err != nil {
		return nil, err
	}

	loader.I("load package -- success %s", time.Now().Sub(start))

	return loader, nil
}

func (loader *Loader) addpackage(domain string, pkg *Package) {
	packages, ok := loader.packages[domain]

	if !ok {
		packages = make(map[string]*Package)

		loader.packages[domain] = packages
	}

	pkg.loadPath = make([]*Package, len(loader.checkerOfDCG))

	copy(pkg.loadPath, loader.checkerOfDCG)

	packages[pkg.Name] = pkg
}

func (loader *Loader) querypackage(domain string, name string) (*Package, bool) {
	if packages, ok := loader.packages[domain]; ok {

		if pkg, ok := packages[name]; ok {
			return pkg, true
		}

	}

	return nil, false
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

	domains := ParseDomain(pkg.Domain, DomainDefault)

	hasTask := false

	for _, v := range domains {
		if v == "task" {
			hasTask = true
			break
		}
	}

	if !hasTask {
		domains = append(domains, "task")
	}

	for _, domain := range domains {

		_, _, err := loader.tryMount(domain, pkg.Name, loader.targetpath, "current", "file")

		if err != nil {
			return err
		}

		pkg, err = loader.loadpackagev2(domain, pkg.Name, loader.targetpath)

		if err != nil {
			return err
		}

		loader.addpackage(domain, pkg)

		if domain == "task" {

			for _, ir := range loader.imports {
				loader.importPackage(domain, pkg, ir)
			}
		}

	}

	loader.D("loaded domain : [%s]", strings.Join(domains, ","))

	// try load gsmake
	if _, ok := loader.querypackage("task", "github.com/gsmake/gsmake"); !ok {

		pkg, err := loader.loadpackage(Import{
			Name:    "github.com/gsmake/gsmake",
			Version: VersionGSMake,
			SCM:     "git",
			Domain:  "task",
		})

		if err != nil {
			return gserrors.Newf(err, "load package github.com/gsmake/gsmake error")
		}

		loader.addpackage("task", pkg)
	}

	// dismount not loaded packages

	err = loader.rootfs.List(func(src, target *vfs.Entry) bool {

		loader.D("check mounted vfs node :%s", target)

		if _, ok := loader.querypackage(target.Domain(), target.Name()); !ok {
			loader.I("dismount :%s", target)
			loader.rootfs.Dismount(target.String())
		}

		return true
	})

	if err != nil {
		return gserrors.Newf(err, "dismount unreference packages error")
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

func (loader *Loader) tryMount(domain, name, srcpath, version, scm string) (string, string, error) {
	target := fmt.Sprintf("gsmake://%s?domain=%s", name, domain)

	src := fmt.Sprintf("%s://%s?version=%s", scm, srcpath, version)

	if !loader.rootfs.Mounted(src, target) {

		if err := loader.rootfs.Mount(src, target); err != nil {
			return src, target, err
		}
	}

	return src, target, nil
}

func loadpath(path []*Package, name, version string) string {
	var buff bytes.Buffer

	for _, pkg := range path {
		buff.WriteString(fmt.Sprintf("\t\t%s %s\n", pkg.Name, pkg.Version))
	}

	buff.WriteString(fmt.Sprintf("\t\t%s %s\n", name, version))

	return buff.String()
}

func (loader *Loader) loadpackage(i Import) (*Package, error) {

	loader.D("  %s %s", i.Name, i.Domain)

	if pkg, ok := loader.querypackage(i.Domain, i.Name); ok {
		if pkg.Version != i.Version {
			return nil, gserrors.Newf(
				ErrLoad,
				"%s import package with diff version\n\tthe one:\n%s\n\tthe other:\n%s",
				i.Domain,
				loadpath(pkg.loadPath, pkg.Name, pkg.Version),
				loadpath(loader.checkerOfDCG, i.Name, i.Version),
			)
		}

		return pkg, nil
	}

	// DCG check
	if err := loader.checkDCG(i.Name); err != nil {
		return nil, err
	}

	_, target, err := loader.tryMount(i.Domain, i.Name, i.Name, i.Version, i.SCM)

	if err != nil {
		return nil, err
	}

	_, entry, err := loader.rootfs.Open(target)

	if err != nil {
		return nil, err
	}

	importpkg, err := loader.loadpackagev2(i.Domain, i.Name, entry.Mapping)

	if err != nil {
		return nil, err
	}

	importpkg.Version = i.Version

	return importpkg, nil
}

func (loader *Loader) loadpackagev2(currentDomain, name, fullpath string) (*Package, error) {

	jsonfile := filepath.Join(fullpath, ".gsmake.json")

	if !fs.Exists(jsonfile) {
		// this package is a traditional golang package
		return &Package{
			Name:   name,
			Domain: currentDomain,
		}, nil
	}

	pkg, err := loadjson(jsonfile)

	if err != nil {
		return nil, err
	}

	// parse redirect instruction
	if pkg.Redirect != nil {

		if pkg.Redirect.Version == "" {
			pkg.Redirect.Version = "current"
		}

		if err := loader.parseSCM(pkg.Redirect); err != nil {
			return nil, err
		}

		loader.I("redirect :\n\tsource :%s %s\n\ttarget :%s %s", pkg.Name, pkg.Version, pkg.Redirect.Name, pkg.Redirect.Version)

		pkg.Redirect.Domain = currentDomain

		return loader.loadpackage(*pkg.Redirect)
	}

	loader.checkerOfDCG = append(loader.checkerOfDCG, pkg)

	defer func() {
		loader.checkerOfDCG = loader.checkerOfDCG[:len(loader.checkerOfDCG)-1]
	}()

	for _, ir := range pkg.Import {

		if err := loader.importPackage(currentDomain, pkg, ir); err != nil {
			return nil, err
		}
	}

	for _, task := range pkg.Task {
		task.Package = name
	}

	return pkg, nil
}

func (loader *Loader) parseSCM(ir *Import) error {
	// calc scm url
	if ir.SCM == "" {

		u, err := url.Parse(fmt.Sprintf("https://%s", ir.Name))

		if err != nil {

			return gserrors.Newf(err, "invalid import package :%s\n%s", ir.Name, loadpath(loader.checkerOfDCG, ir.Name, ir.Version))
		}

		ir.SCM = loader.rootfs.Protocol(u.Host)
	}

	return nil
}

func (loader *Loader) importPackage(currentDomain string, parent *Package, ir Import) error {

	if ir.Version == "" {
		ir.Version = "current"
	}

	// calc scm url
	if err := loader.parseSCM(&ir); err != nil {
		return err
	}

	domains := ParseDomain(ir.Domain, parent.Domain)

	//parentDomains := ParseDomain(parent.Domain, DomainDefault)

	for _, domain := range domains {

		if domain == currentDomain {

			loader.D("%s %s import %s %s", parent.Name, currentDomain, ir.Name, domain)

			ir.Domain = domain

			pkg, err := loader.loadpackage(ir)

			if err != nil {
				return err
			}

			loader.addpackage(domain, pkg)

			return nil
		}

	}

	return nil
}

func loadjson(file string) (*Package, error) {

	content, err := ioutil.ReadFile(file)

	if err != nil {
		return nil, gserrors.Newf(err, "load config file err\n\t%s", file)
	}

	config := &Package{
		Domain: DomainDefault,
	}

	err = json.Unmarshal(content, &config)

	if err != nil {
		return nil, gserrors.Newf(err, "unmarshal .gsmake.json file error\n\tfile:%s", file)
	}

	return config, nil
}
