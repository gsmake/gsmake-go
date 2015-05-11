package gsmake

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gsdocker/gserrors"
)

// gsmake predeined environment variables
const (
	EnvHome = "GSMAKE_HOME"
)

// gsmake consts
const (
	NameConfigFile = ".gsmake.json"
)

// Settings the gsmake settings
type Settings struct {
	home        string // gsmake home path
	repopath    string // gsmake repository path
	workspace   string // gsmake workspace root
	taskpath    string // gsmake task import package root path
	runtimepath string // gsmake runtime import package root path
}

func (settings *Settings) setHome(path string) error {

	if !filepath.IsAbs(path) {
		fullpath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		path = fullpath
	}

	settings.home = path

	settings.repopath = filepath.Join(settings.home, "repo")
	settings.workspace = filepath.Join(settings.home, "workspace")
	settings.taskpath = filepath.Join(settings.home, "task")
	settings.runtimepath = filepath.Join(settings.home, "runtimes")

	return nil
}

func (settings *Settings) clearimport(name string) error {
	if err := settings.cleartaskimport(name); err != nil {
		return err
	}

	if err := settings.clearruntimeimport(name); err != nil {
		return err
	}

	return nil
}

func (settings *Settings) cleartaskimport(name string) error {
	importroot := filepath.Join(settings.taskpath, name, "src")

	if err := os.RemoveAll(importroot); err != nil {
		return gserrors.Newf(err, "can't clear %s import package dir", name)
	}

	return nil
}

func (settings *Settings) clearruntimeimport(name string) error {
	importroot := filepath.Join(settings.runtimepath, name, "src")

	if err := os.RemoveAll(importroot); err != nil {
		return gserrors.Newf(err, "can't clear %s import package dir", name)
	}

	return nil
}

func (settings *Settings) repoPath(name string) string {
	return filepath.Join(settings.repopath, name)
}

func (settings *Settings) devpath(name string) string {
	return filepath.Join(settings.workspace, name, "src", name)
}

func (settings *Settings) devbinpath(name string) string {
	return filepath.Join(settings.workspace, name, "bin")
}

func (settings *Settings) taskPath(name string, importpath string) string {
	return filepath.Join(settings.taskpath, name, "src", importpath)
}

func (settings *Settings) runtimesPath(name string, importpath string) string {
	return filepath.Join(settings.runtimepath, name, "src", importpath)
}

func (settings *Settings) runtimesGOPath(name string) string {
	return fmt.Sprintf("%s%s%s",
		filepath.Join(settings.workspace, name),
		string(os.PathListSeparator),
		filepath.Join(settings.runtimepath, name),
	)
}

func (settings *Settings) taskGOPath(name string) string {
	return fmt.Sprintf("%s%s%s",
		filepath.Join(settings.workspace, name),
		string(os.PathListSeparator),
		filepath.Join(settings.taskpath, name),
	)
}
