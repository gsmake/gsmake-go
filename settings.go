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
	home      string // gsmake home path
	repopath  string // gsmake repository path
	workspace string // gsmake workspace root
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

	return nil
}

func (settings *Settings) clearworkimport(name string) error {
	importroot := filepath.Join(settings.workspace, name, "import")

	if err := os.RemoveAll(importroot); err != nil {
		return gserrors.Newf(err, "can't clear %s import package dir", name)
	}

	return nil
}

func (settings *Settings) worksrcpath(name string) string {
	return filepath.Join(settings.workspace, name, "dev", "src", name)
}

func (settings *Settings) workimportpath(name string, importname string) string {
	return filepath.Join(settings.workspace, name, "import", "src", importname)
}

func (settings *Settings) workgopath(name string) string {
	return fmt.Sprintf(
		"%s%s%s",
		filepath.Join(settings.workspace, name, "dev"),
		string(os.PathListSeparator),
		filepath.Join(settings.workspace, name, "import"),
	)
}
