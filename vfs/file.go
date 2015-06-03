package vfs

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
	"github.com/gsdocker/gsos/fs"
)

// ErrFileFS .
var (
	ErrFileFS = errors.New("file fs error")
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
			return gserrors.Newf(ErrFileFS, "create mount target dir error\n\t%s", dir)
		}
	}

	srcpath := fmt.Sprintf("%s%s", src.Host, src.Path)

	if err := fs.Symlink(srcpath, target.Mapping); err != nil {
		return gserrors.Newf(err, "link dir error\n\tsrc: %s\n\ttarget: %s", srcpath, target.Mapping)
	}

	return nil
}

// UpdateCache implement UserFS
func (fileFS *FileFS) UpdateCache(rootfs RootFS, cachepath string) error {
	return nil
}

// Dismount implement UserFS
func (fileFS *FileFS) Dismount(rootfs RootFS, src, target *Entry) error {

	if fs.Exists(target.Mapping) {
		if err := fs.RemoveAll(target.Mapping); err != nil {
			return gserrors.Newf(ErrFileFS, "remove mount target dir error\n%s", target.Mapping)
		}
	}

	return nil
}

// Update implement UserFS
func (fileFS *FileFS) Update(rootfs RootFS, src, target *Entry, nocache bool) error {
	return fileFS.Mount(rootfs, src, target)
}
