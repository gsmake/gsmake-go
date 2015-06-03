package vfs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
	"github.com/gsdocker/gsos/fs"
	"github.com/gsdocker/gsos/uuid"
)

// Site .
type Site struct {
	SCM     string
	URL     string
	Pattern string
}

// MountIndexer .
type MountIndexer struct {
	Src    string
	Target string
}

// Metadata .
type Metadata struct {
	gslogger.Log
	rootpath   string // gsmake root path
	dbpath     string // Metadata directory
	flocker    string // flock filename
	userspace  string // userspace directory
	targetpath string
}

func newMetadata(rootpath string, targetpath string) (*Metadata, error) {
	db := &Metadata{
		Log:        gslogger.Get("vfs_Metadata"),
		rootpath:   rootpath,
		flocker:    filepath.Join(rootpath, ".db", "locker"),
		dbpath:     filepath.Join(rootpath, ".db"),
		targetpath: targetpath,
	}

	if !fs.Exists(db.dbpath) {
		if err := fs.MkdirAll(db.dbpath, 0755); err != nil {
			return nil, gserrors.Newf(err, "create gsmake root directory error")
		}
	}

	err := fs.FLock(db.flocker, func() error {

		var userspaces map[string]string

		if err := db.readIndexer("userspace", &userspaces); err != nil {
			return err
		}

		// remove temp userspace
		for k, v := range userspaces {
			if strings.HasPrefix(k, os.TempDir()) {

				path := filepath.Join(db.dbpath, v)

				if fs.Exists(path) {

					if err := fs.RemoveAll(path); err != nil {
						return gserrors.Newf(err, "clear temp userspace metadata error\n%s", path)
					}
				}

				path = filepath.Join(rootpath, "userspace", v)

				if fs.Exists(path) {
					if err := fs.RemoveAll(path); err != nil {
						return gserrors.Newf(err, "clear temp userspace error\n%s", path)
					}
				}

			}
		}

		if us, ok := userspaces[targetpath]; ok {
			db.userspace = filepath.Join(rootpath, "userspace", us)
			return nil
		}

		us := uuid.New()

		db.userspace = filepath.Join(rootpath, "userspace", us)

		userspaces[targetpath] = us

		if err := db.writeIndexer("userspace", userspaces); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	err = fs.FLock(db.flocker, func() error {

		var sites map[string]Site

		if err := db.readIndexer("sites", &sites); err != nil {
			return err
		}

		if len(sites) == 0 {
			sites = map[string]Site{
				"github.com": {
					SCM:     "git",
					URL:     "https://${root}.git",
					Pattern: `^(?P<root>github\.com/[A-Za-z0-9_.\-]+/[A-Za-z0-9_.\-]+)(/[A-Za-z0-9_.\-]+)*$`,
				},
				"gopkg.in": {
					SCM:     "git",
					URL:     "https://${root}",
					Pattern: `^(?P<root>gopkg\.in/[A-Za-z0-9_.\-])$`,
				},

				"bitbucket.org": {
					SCM:     "git",
					URL:     "https://${root}.git",
					Pattern: `^(?P<root>bitbucket\.org/[A-Za-z0-9_.\-]+/[A-Za-z0-9_.\-]+)(/[A-Za-z0-9_.\-]+)*$`,
				},
			}
		}

		if err := db.writeIndexer("sites", sites); err != nil {
			return err
		}

		return nil
	})

	return db, err
}

func (db *Metadata) clear() error {

	path := filepath.Join(db.dbpath, filepath.Base(db.userspace))

	if fs.Exists(path) {
		if err := fs.RemoveAll(path); err != nil {
			return gserrors.Newf(err, "remove userspace metadata error\n%s", path)
		}
	}

	return nil
}

func (db *Metadata) queryredirect(from string) (to string, ok bool) {

	fs.FLock(db.flocker, func() error {
		var mapping map[string]string

		if err := db.readIndexer("redirect", &mapping); err != nil {
			ok = false

			return nil
		}

		to, ok = mapping[from]

		return nil

	})

	return
}

func (db *Metadata) redirect(from, to string, enable bool) error {

	return fs.FLock(db.flocker, func() error {

		var mapping map[string]string

		if err := db.readIndexer("redirect", &mapping); err != nil {
			return err
		}

		if enable {
			mapping[from] = to
		} else {
			if old, ok := mapping[from]; ok && old == to {
				delete(mapping, from)
			}
		}

		if err := db.writeIndexer("redirect", mapping); err != nil {
			return err
		}

		return nil
	})
}

func (db *Metadata) site(host string) (site Site, ok bool) {
	fs.FLock(db.flocker, func() error {

		var sites map[string]Site

		if err := db.readIndexer("sites", &sites); err != nil {
			return err
		}

		site, ok = sites[host]

		return nil
	})

	return
}

func (db *Metadata) mountindexer() string {
	return path.Join(filepath.Base(db.userspace), "mount")
}

func (db *Metadata) mountkey(target *Entry) string {
	return fmt.Sprintf("%s/%s%s", target.Query().Get("domain"), target.Host, target.Path)
}

// Mount .
func (db *Metadata) mount(src, target *Entry) error {

	gserrors.Require(target.Scheme == FSGSMake, "target must be rootfs node")

	indexername := db.mountindexer()

	key := db.mountkey(target)

	return db.tx(func() error {

		var indexer map[string]MountIndexer

		if err := db.readIndexer(indexername, &indexer); err != nil {
			return err
		}

		indexer[key] = MountIndexer{src.String(), target.String()}

		if err := db.writeIndexer(indexername, indexer); err != nil {
			return err
		}

		return nil
	})
}

// Dismount .
func (db *Metadata) dismount(src, target *Entry) error {

	gserrors.Require(target.Scheme == FSGSMake, "target must be rootfs node")

	indexername := db.mountindexer()

	key := db.mountkey(target)

	return db.tx(func() error {

		var indexer map[string]MountIndexer

		if err := db.readIndexer(indexername, &indexer); err != nil {
			return err
		}

		delete(indexer, key)

		if err := db.writeIndexer(indexername, indexer); err != nil {
			return err
		}

		return nil
	})
}

func (db *Metadata) queryMount(rootfs *VFS, target *Entry) (entry *Entry, err error) {

	gserrors.Require(target.Scheme == FSGSMake, "target must be rootfs node")

	indexername := db.mountindexer()

	key := db.mountkey(target)

	db.tx(func() error {

		var indexer map[string]MountIndexer

		if err = db.readIndexer(indexername, &indexer); err != nil {
			return nil
		}

		if v, ok := indexer[key]; ok {
			entry, err = rootfs.parseurl(v.Src)
			return nil
		}

		err = gserrors.Newf(ErrNotFound, "mount info not found: %s", target)

		return nil
	})

	return
}

// tx start a transaction
func (db *Metadata) tx(f func() error) error {
	return fs.FLock(db.flocker, f)
}

// readIndexer .
func (db *Metadata) readIndexer(name string, indexer interface{}) error {
	indexerfile := filepath.Join(db.dbpath, name+".id")

	if !fs.Exists(indexerfile) {

		if err := fs.MkdirAll(filepath.Dir(indexerfile), 0755); err != nil {
			return gserrors.Newf(err, "create indexer error")
		}

		if err := ioutil.WriteFile(indexerfile, []byte("{}"), 0755); err != nil {
			return gserrors.Newf(err, "create indexer error")
		}
	}

	content, err := ioutil.ReadFile(indexerfile)

	if err != nil {
		return gserrors.Newf(err, "read %s indexer error", name)
	}

	err = json.Unmarshal(content, indexer)

	if err != nil {
		return gserrors.Newf(err, "read %s indexer error", name)
	}

	return nil
}

// writeIndexer .
func (db *Metadata) writeIndexer(name string, indexer interface{}) error {

	indexerfile := filepath.Join(db.dbpath, name+".id")

	content, err := json.Marshal(indexer)

	if err != nil {
		return gserrors.Newf(err, "marshal %s indexer error", name)
	}

	var fmtjson bytes.Buffer

	json.Indent(&fmtjson, content, "", "\t")

	err = ioutil.WriteFile(indexerfile, fmtjson.Bytes(), 0755)

	if err != nil {
		return gserrors.Newf(err, "write %s indexer error", name)
	}

	return nil
}

func (db *Metadata) cacheRoot(src *Entry) string {
	return filepath.Join(db.rootpath, "cache", src.Scheme, src.Host, src.Path)
}
