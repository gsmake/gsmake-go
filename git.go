package gsmake

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
	"github.com/gsdocker/gsos"
)

// ErrGit .
var (
	ErrGit = errors.New("git command error")
)

type gitSCM struct {
	gslogger.Log        // mixin log APIs
	homepath     string // gsmake home path
	cmd          string // command name
	name         string // command display name
}

func newGitSCM(homepath string) (*gitSCM, error) {
	_, err := SearchCmd("git")

	if err != nil {
		return nil, err
	}

	return &gitSCM{
		Log:      gslogger.Get("gsmake"),
		cmd:      "git",
		name:     "GIT",
		homepath: homepath,
	}, nil
}

func (git *gitSCM) String() string {
	return git.name
}

func (git *gitSCM) Cmd() string {
	return git.cmd
}

// Update implement SCM interface func
func (git *gitSCM) Update(url string, name string, version string) (string, error) {

	repopath := RepoDir(git.homepath, name)

	git.D("package git repo\n\tname:%s\n\tpath:%s", name, repopath)

	if !gsos.IsExist(repopath) {
		return "", gserrors.Newf(ErrGit, "package repo not exist\n\tname:%s\n\tpath:%s", name, repopath)
	}

	info, err := os.Lstat(repopath)

	if err != nil {
		return "", gserrors.Newf(err, "read repo dir info error")
	}

	if info.Mode()&os.ModeSymlink == os.ModeSymlink {
		git.I("package[%s] git pull -- skipped, cached developing package", name)
	}

	cmd := exec.Command("git", "pull", "--all")

	cmd.Dir = repopath

	var buff bytes.Buffer

	cmd.Stderr = &buff

	git.I("package[%s] git update --all ", name)

	err = cmd.Run()

	if err != nil {
		return "", gserrors.Newf(err, "err call :git pull\n%s", buff.String())
	}

	return repopath, nil
}

func (git *gitSCM) Create(url string, name string, version string) (string, error) {

	repopath := RepoDir(git.homepath, name)

	git.D("package git repo\n\tname:%s\n\tpath:%s", name, repopath)

	if !gsos.IsExist(repopath) {

		if err := os.MkdirAll(filepath.Dir(repopath), 0755); err != nil {
			return "", gserrors.Newf(err, "create package repo error")
		}

		cmd := exec.Command("git", "clone", url)

		var buff bytes.Buffer

		cmd.Stderr = &buff

		cmd.Dir = filepath.Dir(repopath)

		git.I("package[%s] git clone %s", name, url)

		err := cmd.Run()

		if err != nil {
			return "", gserrors.Newf(err, "err call :git clone %s\n%s", url, buff.String())
		}
	}

	return repopath, nil
}

func (git *gitSCM) Copy(name string, version string, targetpath string) error {

	repopath := RepoDir(git.homepath, name)

	git.D("package git repo\n\tname:%s\n\tpath:%s", name, repopath)

	if !gsos.IsExist(repopath) {
		return gserrors.Newf(ErrGit, "package repo not exist\n\tname:%s\n\tpath:%s", name, repopath)
	}

	if gsos.IsExist(targetpath) {
		if err := gsos.RemoveAll(targetpath); err != nil {
			return gserrors.Newf(err, "remove exist target directory error")
		}
	}

	if err := gsos.CopyDir(repopath, targetpath); err != nil {
		return gserrors.Newf(err, "copy cached package to target directory error")
	}

	// chdir to target directory and execute git checkout command

	if version == "current" {
		version = "master"
	}

	cmd := exec.Command("git", "checkout", version)

	var buff bytes.Buffer

	cmd.Stderr = &buff

	cmd.Dir = targetpath

	err := cmd.Run()

	if err != nil {
		return gserrors.Newf(err, "git checkout\n%s", buff.String())
	}

	return nil
}

// Cache cache package as global repo package
func (git *gitSCM) Cache(name string, version string, source string) error {

	repopath := RepoDir(git.homepath, name)

	if gsos.SameFile(repopath, source) {
		return nil
	}

	if err := gsos.RemoveAll(repopath); err != nil {
		return gserrors.Newf(err, "remove global repo dir error")
	}

	if err := os.MkdirAll(filepath.Dir(repopath), 0755); err != nil {
		return gserrors.Newf(err, "create global repo dir error")
	}

	if err := gsos.Symlink(source, repopath); err != nil {
		return gserrors.Newf(err, " cache package to global repo error")
	}

	return nil
}

// RemoveCache remove cached package
func (git *gitSCM) RemoveCache(name string, version string, source string) error {

	repopath := RepoDir(git.homepath, name)

	if !gsos.SameFile(repopath, source) {
		return nil
	}

	if err := gsos.RemoveAll(repopath); err != nil {
		return gserrors.Newf(err, "remove cached repo dir error")
	}

	return nil

}
