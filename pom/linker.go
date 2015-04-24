package pom

import (
	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
	"github.com/gsdocker/gsos"
)

// Linker project linker
type Linker struct {
	gslogger.Log                     // mixin log APIs
	rootPath     string              // gsmake root path
	local        map[string]*Project // local projects
}

func newLinker(rootPath string) (linker *Linker, err error) {

	if !gsos.IsDir(rootPath) {
		return nil, gserrors.Newf(ErrLinker, "gsmake root path not exist : \n\t%s", rootPath)
	}

	return &Linker{
		Log:      gslogger.Get("gsmake"),
		rootPath: rootPath,
		local:    make(map[string]*Project),
	}, nil
}

func (linker *Linker) registerLocal(project *Project) {
	linker.local[project.Name] = project
}

func (linker *Linker) link(importer *Project, name string, version string) (string, bool) {
	if project, ok := linker.local[name]; ok {
		if project.Version == version {
			return project.path, true
		}
	}

	return "", false
}
