package gsmake

import (
	"fmt"
	"os"
	"path/filepath"
)

// gsmake predeined environment variables
const (
	EnvHome = "GSMAKE_HOME"
)

// gsmake consts
const (
	NameConfigFile = ".gsmake.json"
)

// BinaryDir calc package binary generate directory
func BinaryDir(homepath string, packagename string) string {
	return filepath.Join(Workspace(homepath, packagename), "bin")
}

// RepoDir calc package repo directory
func RepoDir(homepath string, packagename string) string {
	return filepath.Join(homepath, "repo", packagename)
}

// Workspace calc package workspace directory
func Workspace(homepath string, packagename string) string {
	return filepath.Join(homepath, "workspace", packagename)
}

// WorkspaceImportDir .
func WorkspaceImportDir(homepath string, packagename string) string {
	return filepath.Join(homepath, "workspace", packagename, "src", packagename)
}

// TaskStageImportDir calc task stage import package root directory
func TaskStageImportDir(homepath string, packagename string, importpath string) string {
	return filepath.Join(homepath, "task", packagename, "src", importpath)
}

// RuntimesStageImportDir calc runtimes stage import package root directory
func RuntimesStageImportDir(homepath string, packagename string, importpath string) string {
	return filepath.Join(homepath, "runtimes", packagename, "src", importpath)
}

// RuntimesStageGOPATH calc runtimes stage GOPATH env value
func RuntimesStageGOPATH(homepath string, packagename string) string {
	return fmt.Sprintf(
		"%s%s%s",
		filepath.Join(homepath, "workspace", packagename),
		string(os.PathListSeparator),
		filepath.Join(homepath, "runtimes", packagename),
	)
}

// TaskStageGOPATH calc task stage GOPATH env value
func TaskStageGOPATH(homepath string, packagename string) string {
	return fmt.Sprintf(
		"%s%s%s",
		filepath.Join(homepath, "workspace", packagename),
		string(os.PathListSeparator),
		filepath.Join(homepath, "task", packagename),
	)
}
