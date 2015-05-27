package vfs

import (
	"errors"
	"path/filepath"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
	"github.com/gsmake/gsmake/fs"
)

// ErrGitFS .
var (
	ErrGitFS = errors.New("git fs error")
)

// GitFS git fs for gsmake vfs
type GitFS struct {
	gslogger.Log // Mixin log APIs
}

// Mount implement UserFS
func (gitFS *GitFS) String() string {
	return "git"
}

// Mount implement UserFS
func (gitFS *GitFS) Mount(rootfs RootFS, src, target *Entry) error {

	cachepath := filepath.Join(rootfs.Metadata().CacheRoot(), "git", src.Host, src.Path)

	cachepath = filepath.Dir(cachepath)

	if err := fs.MkdirAll(cachepath, 0755); err != nil {
		return gserrors.Newf(ErrGitFS, "mdir cache directory error")
	}

	remote := src.Query().Get("remote")

	if remote == "" {
		return gserrors.Newf(ErrGitFS, "expect remoet url \n%s", src)
	}

	gitFS.D("mount remote url :%s", remote)

	//cmd := exec.Command("git", "clone")

	return nil
}

// Dismount implement UserFS
func (gitFS *GitFS) Dismount(rootfs RootFS, src, target *Entry) error {
	return nil
}

// Update implement UserFS
func (gitFS *GitFS) Update(rootfs RootFS, src, target *Entry) error {
	gserrors.Require(target.Scheme == FSGSMake, "target must be rootfs node")
	gserrors.Require(target.Scheme == "git", "src must be gitfs node")

	return nil
}

// Commit implement UserFS
func (gitFS *GitFS) Commit(rootfs RootFS, src, target *Entry) error {
	return nil
}

// UpdateAll implement UserFS
func (gitFS *GitFS) UpdateAll(rootfs RootFS) error {
	return nil
}
