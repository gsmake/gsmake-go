package vfs

import (
	"errors"
	"net/url"
	"path/filepath"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
	"github.com/gsmake/gsmake/fs"
)

// errors
var (
	ErrURL = errors.New("error url")
)

// builtin domain
const (
	DomainTask     = "task"
	VersionDefault = "current"
	FSGSMake       = "gsmake"
)

// rootfService vfs rootfService
type rootfService struct {
	gslogger.Log                   // Mixin Log APIs
	userspace    string            // userspace path
	rootpath     string            // gsmake root path
	meta         *Metadata         // mixin metadb
	userfs       map[string]UserFS // register userfs
}

// New create new vfs rootfService
func New(rootpath string, username string) (RootFS, error) {

	log := gslogger.Get("vfs")

	log.I("init vfs ...")

	fullpath, err := filepath.Abs(rootpath)

	if err != nil {
		return nil, err
	}

	db, err := newMetadata(fullpath, username)

	if err != nil {
		return nil, err
	}

	s := &rootfService{
		meta:      db,
		Log:       log,
		userspace: db.userspace,
		rootpath:  fullpath,
		userfs:    make(map[string]UserFS),
	}

	s.D("rootpath :%s", s.rootpath)
	s.D("userspace :%s", s.userspace)

	if !fs.Exists(s.userspace) {
		if err := fs.MkdirAll(s.userspace, 0755); err != nil {
			return nil, gserrors.Newf(err, "create userspace directory error\n%s", s.userspace)
		}
	}

	s.I("init vfs -- success")

	return s, nil
}

func (rootfs *rootfService) parseurl(src string) (*Entry, error) {

	u, err := url.Parse(src)

	if err != nil {
		return nil, gserrors.Newf(err, "parse url error ")
	}

	entry := &Entry{
		URL: u,
	}

	if entry.Query().Get("domain") == "" {
		entry.Query().Set("domain", DomainTask)
	}

	if entry.Query().Get("version") == "" {
		entry.Query().Set("version", VersionDefault)
	}

	if entry.Scheme != FSGSMake {
		userfs, ok := rootfs.userfs[entry.Scheme]

		if !ok {
			return nil, gserrors.Newf(ErrURL, "unknown fs type :%s", entry.Scheme)
		}

		entry.userfs = userfs
	} else {
		entry.Mapping = filepath.Join(rootfs.userspace, entry.Query().Get("domain"), "src", entry.Host, entry.Path)
	}

	return entry, nil
}

// implement RootFS interface
func (rootfs *rootfService) Mount(src, target string) error {

	if err := rootfs.Dismount(target); err != nil {
		return err
	}

	srcE, err := rootfs.parseurl(src)

	if err != nil {
		return err
	}

	targetE, err := rootfs.parseurl(target)

	if err != nil {
		return err
	}

	if srcE.userfs == nil {
		return gserrors.Newf(ErrURL, "mount source can't be vfs url\n%s", src)
	}

	if targetE.Scheme != FSGSMake {
		return gserrors.Newf(ErrURL, "mount target must be vfs url\n%s", src)
	}

	if err := srcE.userfs.Mount(rootfs, srcE, targetE); err != nil {
		return err
	}

	return rootfs.meta.Mount(srcE, targetE)
}

// implement RootFS interface
func (rootfs *rootfService) Dismount(target string) error {

	targetE, err := rootfs.parseurl(target)

	if err != nil {
		return err
	}

	if targetE.Scheme != FSGSMake {
		return gserrors.Newf(ErrURL, "mount target must be vfs url\n%s", target)
	}

	srcE, err := rootfs.meta.queryMount(rootfs, targetE)

	if err != nil || srcE == nil {
		return err
	}

	if err := srcE.userfs.Dismount(rootfs, srcE, targetE); err != nil {
		return err
	}

	return rootfs.meta.Dismount(srcE, targetE)
}

// implement RootFS interface
func (rootfs *rootfService) Update(target string) error {

	targetE, err := rootfs.parseurl(target)

	if err != nil {
		return err
	}

	srcE, err := rootfs.meta.queryMount(rootfs, targetE)

	if err != nil {
		return err
	}

	if srcE == nil {
		return gserrors.Newf(ErrURL, "target package not exists \n%s", target)
	}

	return srcE.userfs.Update(rootfs, srcE, targetE)
}

// implement RootFS interface
func (rootfs *rootfService) Commit(target string) error {

	targetE, err := rootfs.parseurl(target)

	if err != nil {
		return err
	}

	srcE, err := rootfs.meta.queryMount(rootfs, targetE)

	if err != nil {
		return err
	}

	if srcE == nil {
		return gserrors.Newf(ErrURL, "target package not exists \n%s", target)
	}

	return srcE.userfs.Commit(rootfs, srcE, targetE)
}

// implement RootFS interface
func (rootfs *rootfService) Metadata() *Metadata {
	return rootfs.meta
}
