package gsmake

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
	"github.com/gsdocker/gsos"
)

// VCSCmd A vcsCmd describes how to use a version control system
// like Mercurial, Git, or Subversion.
type VCSCmd interface {
	// Mxin stringer
	fmt.Stringer
	// vcs command name
	Cmd() string
	// vcs url suffix
	Suffix() string
	// Create command to download a fresh copy of a repository
	Create(properties Properties) error
	// Update command to download updates into an existing repository
	Update(properties Properties) error
}

// Prepare check command env
func Prepare(cmd VCSCmd) error {
	_, err := exec.LookPath(cmd.Cmd())
	if err != nil {
		return gserrors.Newf(
			err,
			"go: missing %s command. See http://golang.org/s/gogetcmd\n",
			cmd)
	}

	return nil
}

// GitCmd Git VCSCmd implement.
type GitCmd struct {
	gslogger.Log        // mixin Log APIs
	cmd          string // command program name
	name         string // command display name
	suffix       string // vcs url suffix
}

// NewGitCmd .
func NewGitCmd() *GitCmd {
	return &GitCmd{
		Log:    gslogger.Get("gsmake"),
		cmd:    "git",
		name:   "Git",
		suffix: ".git",
	}
}

func (cmd *GitCmd) String() string {
	return cmd.name
}

// Cmd  implement VCSCmd interface
func (cmd *GitCmd) Cmd() string {
	return cmd.cmd
}

// Suffix implement VCSCmd interface
func (cmd *GitCmd) Suffix() string {
	return cmd.suffix
}

// Create implement VCSCmd interface
func (cmd *GitCmd) Create(properties Properties) error {
	//first clone all
	if err := Prepare(cmd); err != nil {
		return err
	}

	dir := properties["dir"].(string)

	if gsos.IsExist(dir) {
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
	}

	repo := properties["repo"].(string)

	command := exec.Command(cmd.name, "clone", repo, dir)

	cmd.D("git clone\n\trepo :%s\n\tdir :%s", repo, dir)

	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	if err := command.Run(); err != nil {
		return err
	}

	if _, ok := properties["version"]; !ok {
		properties["version"] = "master"
	}

	if properties["version"] == "current" {
		properties["version"] = "master"
	}

	currentDir := gsos.CurrentDir()

	if err := os.Chdir(dir); err != nil {
		return err
	}

	defer func() {
		os.Chdir(currentDir)
	}()

	command = exec.Command(cmd.name, "checkout", properties["version"].(string))

	cmd.D("git checkout %s", properties["version"].(string))

	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	return command.Run()
}

// Update implement VCSCmd interface
func (cmd *GitCmd) Update(properties Properties) error {
	//first clone all
	if err := Prepare(cmd); err != nil {
		return err
	}

	dir := properties["dir"].(string)

	currentDir := gsos.CurrentDir()

	if err := os.Chdir(dir); err != nil {
		return err
	}

	defer func() {
		os.Chdir(currentDir)
	}()

	command := exec.Command(cmd.name, "pull")

	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	return command.Run()
}
