package vfs

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
	"github.com/gsdocker/gsos/fs"
	"github.com/gsmake/gsmake/property"
)

// errors
var (
	ErrURL      = errors.New("error url")
	ErrFS       = errors.New("unknown fs")
	ErrNotFound = errors.New("vfs node not found")
)

// Exists .
func Exists(rootfs RootFS, target string) bool {
	_, _, err := rootfs.Open(target)

	return err == nil
}

// NotFound .
func NotFound(err error) bool {
	for {
		if gserror, ok := err.(gserrors.GSError); ok {
			err = gserror.Origin()
			continue
		}

		break
	}

	if err == ErrNotFound {
		return true
	}

	return false
}

// builtin domain
const (
	DomainTask     = "task"
	VersionDefault = "current"
	FSGSMake       = "gsmake"
	FSFile         = "file"
)

// Entry filesystem entry
type Entry struct {
	*url.URL
	Mapping string // mapping os path
	userfs  UserFS // userfs object
}

func (entry *Entry) String() string {
	return entry.URL.String()
}

// Domain .
func (entry *Entry) Domain() string {
	return entry.Query().Get("domain")
}

// Name .
func (entry *Entry) Name() string {
	return fmt.Sprintf("%s%s", entry.Host, entry.Path)
}

// RootFS the gsmake vfs rootfs object
type RootFS interface {
	// Mixin Log APIs
	gslogger.Log
	// RootPath .
	RootPath() string
	// TargetPath .
	TargetPath() string
	// Userspace .
	Userspace() string
	// Mounted check if src had mounted on to target
	Mounted(src string, target string) bool
	//Mount mount src fs on to rootfs node
	Mount(src, target string) error
	//Dismount dismount src fs from rootfs node
	Dismount(target string) error
	// List list all mounted nodes
	List(f func(src *Entry, target *Entry) bool) error
	// Open open vfs node, the return values are src entry and target entry
	Open(url string) (src *Entry, target *Entry, err error)
	// Update update a package
	Update(src string, nocache bool) error
	// UpdateAll update all userspace's packages
	UpdateAll(nocache bool) error
	// UpdateCache
	UpdateCache(src string) error
	// Clear clear userspace
	Clear() error
	// Get mount fs cache root
	CacheRoot(src *Entry) string
	// Cached update cache metadata
	Cached(src *Entry) error
	// Protocol get host default protocol
	Protocol(host string) string
	// TempDir domain tempdir
	TempDir(domain string) string
	// DomainDir domain root dir
	DomainDir(domain string) string
	// Redirect set redirect flag
	Redirect(from, to string, enable bool) error
}

//UserFS .
type UserFS interface {
	fmt.Stringer
	// Mount mount userfs on to rootfs
	// @param target The target
	Mount(rootfs RootFS, src *Entry, target *Entry) error
	//Dismount dismount src fs from rootfs node
	Dismount(rootfs RootFS, src, target *Entry) error
	// Update update a package
	Update(rootfs RootFS, src *Entry, target *Entry, nocache bool) error
	// UpdateCache .
	UpdateCache(rootfs RootFS, cachepath string) error
}

// VFS vfs VFS
type VFS struct {
	gslogger.Log                   // Mixin Log APIs
	userspace    string            // userspace path
	rootpath     string            // gsmake root path
	targetpath   string            // targetpath
	meta         *Metadata         // mixin metadb
	userfs       map[string]UserFS // register userfs
}

// New create new vfs VFS
func New(rootpath string, targetpath string) (RootFS, error) {
	rootfs, err := createRootFS(rootpath, targetpath)

	if err != nil {
		return nil, err
	}

	rootfs.userfs = map[string]UserFS{
		"git":  NewGitFS(),
		"file": NewFileFS(),
	}

	return rootfs, nil
}

func createRootFS(rootpath string, targetpath string) (*VFS, error) {
	log := gslogger.Get("vfs")

	log.D("init vfs ...")

	fullpath, err := filepath.Abs(rootpath)

	if err != nil {
		return nil, gserrors.Newf(err, "get abs path error\n\t%s", rootpath)
	}

	targetfullpath, err := filepath.Abs(targetpath)

	if err != nil {
		return nil, gserrors.Newf(err, "get abs path error\n\t%s", targetpath)
	}

	db, err := newMetadata(fullpath, targetfullpath)

	if err != nil {
		return nil, err
	}

	s := &VFS{
		meta:       db,
		Log:        log,
		userspace:  db.userspace,
		rootpath:   fullpath,
		targetpath: targetfullpath,
	}

	s.D("rootpath :%s", s.rootpath)
	s.D("userspace :%s", s.userspace)

	if !fs.Exists(s.userspace) {
		if err := fs.MkdirAll(s.userspace, 0755); err != nil {
			return nil, gserrors.Newf(err, "create userspace directory error\n%s", s.userspace)
		}
	}

	s.D("init vfs -- success")

	return s, nil
}

func (rootfs *VFS) parseurl(src string) (*Entry, error) {

	u, err := url.Parse(src)

	if err != nil {
		return nil, gserrors.Newf(err, "parse url error ")
	}

	entry := &Entry{
		URL: u,
	}

	switch entry.Scheme {
	case FSGSMake:
		if entry.Query().Get("domain") == "" {
			return nil, gserrors.Newf(ErrURL, "expect domain :%s", src)
		}

		entry.Mapping = filepath.Join(rootfs.userspace, entry.Query().Get("domain"), "src", entry.Host, entry.Path)

	case FSFile:

		userfs, ok := rootfs.userfs[entry.Scheme]

		if !ok {
			return nil, gserrors.Newf(ErrURL, "unknown fs type :%s", entry.Scheme)
		}

		entry.userfs = userfs

	default:
		if entry.Query().Get("version") == "" {
			return nil, gserrors.Newf(ErrURL, "expect version :%s", src)
		}

		if entry.Query().Get("remote") == "" {

			site, ok := rootfs.meta.site(entry.Host)

			if !ok {
				return nil, gserrors.Newf(ErrURL, "expect remote url :%s", src)
			}

			matcher, err := regexp.Compile(site.Pattern)

			if err != nil {
				return nil, gserrors.Newf(err, "compile site(%s) import path regexp error", entry.Host)
			}

			url := fmt.Sprintf("%s%s", entry.Host, entry.Path)

			m := matcher.FindStringSubmatch(url)

			if m == nil {

				return nil, gserrors.Newf(ErrURL, "expect remote url :%s", src)
			}

			// Build map of named subexpression matches for expand.
			properties := property.Properties{}

			for i, name := range matcher.SubexpNames() {

				if name != "" && properties[name] == nil {
					properties[name] = m[i]
				}
			}

			if entry.RawQuery == "" {
				entry.RawQuery = "remote=" + properties.Expand(site.URL)
			} else {
				entry.RawQuery = entry.RawQuery + "&remote=" + properties.Expand(site.URL)
			}
		}

		userfs, ok := rootfs.userfs[entry.Scheme]

		if !ok {
			return nil, gserrors.Newf(ErrURL, "unknown fs type :%s", entry.Scheme)
		}

		entry.userfs = userfs
	}

	return entry, nil
}

// RootPath .
func (rootfs *VFS) RootPath() string {
	return rootfs.rootpath
}

// TargetPath .
func (rootfs *VFS) TargetPath() string {
	return rootfs.targetpath
}

// Userspace .
func (rootfs *VFS) Userspace() string {
	return rootfs.userspace
}

// Mount implement RootFS interface
func (rootfs *VFS) Mount(src, target string) error {

	rootfs.D("mount src : %s", src)

	if to, ok := rootfs.meta.queryredirect(src); ok {

		rootfs.D("redirect \n\tfrom :%s\n\tto :%s", src, to)

		src = to
	}

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

	return rootfs.meta.mount(srcE, targetE)
}

// Dismount implement RootFS interface
func (rootfs *VFS) Dismount(target string) error {

	targetE, err := rootfs.parseurl(target)

	if err != nil {
		return err
	}

	if targetE.Scheme != FSGSMake {
		return gserrors.Newf(ErrURL, "mount target must be vfs url\n%s", target)
	}

	srcE, err := rootfs.meta.queryMount(rootfs, targetE)

	if err != nil {

		if NotFound(err) {
			return nil
		}

		return err
	}

	if srcE == nil {
		return nil
	}

	if err := srcE.userfs.Dismount(rootfs, srcE, targetE); err != nil {
		return err
	}

	return rootfs.meta.dismount(srcE, targetE)
}

// Update implement RootFS interface
func (rootfs *VFS) Update(target string, nocache bool) error {

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

	src := fmt.Sprintf("%s://%s%s?version=%s", srcE.Scheme, srcE.Host, srcE.Path, srcE.Query().Get("version"))

	rootfs.D("update src :%s", src)

	if to, ok := rootfs.meta.queryredirect(src); ok {

		rootfs.D("redirect \n\tfrom :%s\n\tto :%s", src, to)

		srcE, err = rootfs.parseurl(to)

		if err != nil {
			return err
		}
	}

	return srcE.userfs.Update(rootfs, srcE, targetE, nocache)
}

// Cached implement RootFS interface
func (rootfs *VFS) Cached(src *Entry) error {

	return rootfs.meta.tx(func() error {

		indexername := "cached"

		var indexer map[string][2]string

		key := fmt.Sprintf("%s://%s%s", src.Scheme, src.Host, src.Path)

		if err := rootfs.meta.readIndexer(indexername, &indexer); err != nil {
			return err
		}

		if _, ok := indexer[key]; !ok {
			indexer[key] = [2]string{src.userfs.String(), rootfs.meta.cacheRoot(src)}
		}

		if err := rootfs.meta.writeIndexer(indexername, indexer); err != nil {
			return err
		}

		return nil
	})
}

// CacheRoot implement RootFS interface
func (rootfs *VFS) CacheRoot(src *Entry) string {

	return rootfs.meta.cacheRoot(src)
}

// Clear implement RootFS
func (rootfs *VFS) Clear() error {

	err := rootfs.meta.clear()

	if err != nil {
		return err
	}

	return nil
}

// Open implement rootfs
func (rootfs *VFS) Open(target string) (*Entry, *Entry, error) {

	targetE, err := rootfs.parseurl(target)

	if err != nil {
		return nil, nil, err
	}

	if targetE.Scheme != FSGSMake {
		return nil, nil, gserrors.Newf(ErrURL, "open target must be vfs url\n%s", target)
	}

	srcE, err := rootfs.meta.queryMount(rootfs, targetE)

	return srcE, targetE, err
}

// List implement rootfs
func (rootfs *VFS) List(f func(src *Entry, target *Entry) bool) (err error) {

	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	var indexer map[string]MountIndexer

	err = rootfs.meta.tx(func() error {

		indexername := rootfs.meta.mountindexer()

		if err := rootfs.meta.readIndexer(indexername, &indexer); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	for _, v := range indexer {

		var (
			src    *Entry
			target *Entry
		)

		src, err = rootfs.parseurl(v.Src)

		if err != nil {
			return err
		}

		target, err = rootfs.parseurl(v.Target)

		if err != nil {
			return err
		}

		if !f(src, target) {
			return
		}
	}

	return
}

// Protocol implement rootfs
func (rootfs *VFS) Protocol(host string) string {

	if site, ok := rootfs.meta.site(host); ok {
		return site.SCM
	}

	return "git"
}

// Mounted implement rootfs
func (rootfs *VFS) Mounted(src string, target string) bool {

	targetE, _, err := rootfs.Open(target)

	if err != nil {
		return false
	}

	srcE, err := rootfs.parseurl(src)

	if err != nil {
		return false
	}

	if srcE.String() == targetE.String() {
		return true
	}

	return false

}

// TempDir .
func (rootfs *VFS) TempDir(domain string) string {
	return filepath.Join(rootfs.userspace, domain, "tmp")
}

// DomainDir .
func (rootfs *VFS) DomainDir(domain string) string {
	return filepath.Join(rootfs.userspace, domain)
}

// Redirect implement rootfs
func (rootfs *VFS) Redirect(from, to string, enable bool) error {

	fromE, err := rootfs.parseurl(from)

	if err != nil {
		return err
	}

	if fromE.Scheme == FSGSMake {
		return gserrors.Newf(ErrURL, "redirect url can't be gsmake://...\n\targ:%s", from)
	}

	toE, err := rootfs.parseurl(to)

	if err != nil {
		return err
	}

	if toE.Scheme == FSGSMake {
		return gserrors.Newf(ErrURL, "redirect url can't be gsmake://...\n\targ:%s", to)
	}

	return rootfs.meta.redirect(from, to, enable)
}

// UpdateCache .
func (rootfs *VFS) UpdateCache(name string) error {

	var indexer map[string][2]string

	err := rootfs.meta.tx(func() error {

		indexername := "cached"

		if err := rootfs.meta.readIndexer(indexername, &indexer); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	if src, ok := indexer[name]; ok {

		userfs, ok := rootfs.userfs[src[0]]

		if !ok {
			return gserrors.Newf(ErrFS, "unknown userfs :%s", src[0])
		}

		err := userfs.UpdateCache(rootfs, src[1])

		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateAll implement rootfs
func (rootfs *VFS) UpdateAll(nocache bool) (err error) {

	var indexer map[string][2]string

	if nocache {

		err := rootfs.meta.tx(func() error {

			indexername := "cached"

			if err := rootfs.meta.readIndexer(indexername, &indexer); err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			return err
		}

		for _, src := range indexer {
			userfs, ok := rootfs.userfs[src[0]]

			if !ok {
				return gserrors.Newf(ErrFS, "unknown userfs :%s", src[0])
			}

			err := userfs.UpdateCache(rootfs, src[1])

			if err != nil {
				return err
			}
		}

	}

	if strings.HasPrefix(rootfs.targetpath, os.TempDir()) {

		rootfs.D("skip update userspace")

		return
	}

	return rootfs.List(func(src, target *Entry) bool {

		err = rootfs.Update(target.String(), false)

		if err != nil {
			return false
		}

		return true
	})
}
