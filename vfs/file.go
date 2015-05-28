package vfs

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
	"github.com/gsmake/gsmake/fs"
)

// ErrFileFS .
var (
	ErrFileFS = errors.New("git fs error")
)

// FileFS git fs for gsmake vfs
type FileFS struct {
	gslogger.Log // Mixin log APIs
}

// NewFileFS create new gitfs system
func NewFileFS() *FileFS {
	return &FileFS{
		Log: gslogger.Get("filefs"),
	}
}

// Mount implement UserFS
func (fileFS *FileFS) String() string {
	return "file"
}

// Mount implement UserFS
func (fileFS *FileFS) Mount(rootfs RootFS, src, target *Entry) error {

	if fs.Exists(target.Mapping) {
		if err := fs.RemoveAll(target.Mapping); err != nil {
			return gserrors.Newf(ErrFileFS, "remove mount target dir error\n%s", target.Mapping)
		}
	}

	dir := filepath.Dir(target.Mapping)

	if !fs.Exists(dir) {
		if err := fs.MkdirAll(dir, 0755); err != nil {
			return gserrors.Newf(ErrFileFS, "create mount target dir error\n%s", dir)
		}
	}

	if err := fs.Symlink(fmt.Sprintf("%s%s", src.Host, src.Path), target.Mapping); err != nil {
		return gserrors.Newf(ErrFileFS, "link dir error\n%s", dir)
	}

	return nil
}

// Dismount implement UserFS
func (fileFS *FileFS) Dismount(rootfs RootFS, src, target *Entry) error {
	return nil
}

// Update implement UserFS
func (fileFS *FileFS) Update(rootfs RootFS, src, target *Entry, nocache bool) error {
	return nil
}
