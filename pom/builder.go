package pom

import (
	"os"
	"path/filepath"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
	"github.com/gsdocker/gsos"
)

// Builder The gsmake builder object
type Builder struct {
	gslogger.Log            // Minx logger
	project        *Project // root project object
	buildDir       string   // build directory
	gsmakeBuildDir string   // gsmake builder build directory
	gsmakeDir      string   // gsmake source code dir
}

// NewBuilder create new builder for target project indicate by param path
func NewBuilder(path string) (*Builder, error) {

	builder := &Builder{
		Log: gslogger.Get("gsmake"),
	}

	var err error

	builder.project, err = builder.createProject(path)

	if err != nil {
		return nil, err
	}

	if buildDir, ok := builder.project.Properties["build.directory"]; ok {
		builder.buildDir = filepath.Join(builder.project.path, buildDir)
	} else {
		builder.buildDir = filepath.Join(path, ".build")
	}

	builder.gsmakeBuildDir = filepath.Join(builder.buildDir, "gsmake")

	if gsmakeDir, ok := builder.project.Properties["gsmake.directory"]; ok {
		builder.gsmakeDir = filepath.Join(builder.project.path, gsmakeDir)
	} else {
		builder.gsmakeDir = filepath.Join(path, ".gsmake")
	}

	builder.I("============project settings============")
	builder.I("= name :%s", builder.project.Name)
	builder.I("= version :%s", builder.project.Version)
	builder.I("= rootDir :%s", builder.project.path)
	builder.I("= buildDir :%s", builder.buildDir)
	builder.I("= gsmakeDir :%s", builder.gsmakeDir)
	builder.I("========================================")

	if !gsos.IsExist(builder.gsmakeBuildDir) {
		err = os.MkdirAll(builder.gsmakeBuildDir, 0755)
		if err != nil {
			return nil, gserrors.Newf(err, "create project build directory error : \n\t%s", builder.gsmakeBuildDir)
		}
	}

	if !gsos.IsExist(builder.gsmakeDir) {
		err = os.MkdirAll(builder.gsmakeDir, 0755)
		if err != nil {
			return nil, gserrors.Newf(err, "create project gsmake directory error : \n\t%s", builder.gsmakeDir)
		}
	}

	return builder, err
}

func (builder *Builder) linkDir(source string, target string) error {

	if !gsos.IsExist(target) {

		dir := filepath.Dir(target)

		err := os.MkdirAll(dir, 0755)

		if err != nil {
			return gserrors.Newf(err, "create directory error \n\t%s", dir)
		}

		err = os.Symlink(source, target)

		if err != nil {
			return gserrors.Newf(err, "create symlink error\n\tsource :%s\n\ttarget :%s\n\t", source, target)
		}

	} else {

		for {
			if t, err := os.Readlink(target); err == nil {
				if t == source {
					break
				}
			}

			if err := os.Remove(target); err != nil {
				return gserrors.Newf(err, "remove directory error \n\t%s", target)
			}

			err := os.Symlink(source, target)

			if err != nil {
				return gserrors.Newf(err, "create symlink error\n\tsource :%s\n\ttarget :%s\n\t", source, target)
			}

			break
		}

	}

	return nil
}

// linkProject link the gsmake project in build directory
func (builder *Builder) linkProject(project *Project) error {

	builder.I("link project %s:%s", project.Name, project.Version)

	linkTarget := filepath.Join(builder.buildDir, "src", project.Name)

	if err := builder.linkDir(builder.gsmakeDir, linkTarget); err != nil {

		builder.I("link project -- failed")

		return err
	}

	builder.I("link project -- finish")

	return nil
}

// Compile compile and creat build program for current project
func (builder *Builder) Compile() error {

	return builder.linkProject(builder.project)
}
