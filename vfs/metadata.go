package vfs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
	"github.com/gsdocker/gsos/uuid"
	"github.com/gsmake/gsmake/fs"
)

// Metadata .
type Metadata struct {
	gslogger.Log
	rootpath  string // gsmake root path
	dbpath    string // Metadata directory
	flocker   string // flock filename
	userspace string // userspace directory
}

func newMetadata(rootpath string, username string) (*Metadata, error) {
	db := &Metadata{
		Log:      gslogger.Get("vfs_Metadata"),
		rootpath: rootpath,
		flocker:  filepath.Join(rootpath, ".db", "locker"),
		dbpath:   filepath.Join(rootpath, ".db"),
	}

	if !fs.Exists(db.dbpath) {
		if err := fs.MkdirAll(db.dbpath, 0755); err != nil {
			return nil, gserrors.Newf(err, "create gsmake root directory error")
		}
	}

	err := fs.FLock(db.flocker, func() error {

		var userspaces map[string]string

		if err := db.ReadIndexer("userspace", &userspaces); err != nil {
			return err
		}

		if us, ok := userspaces[username]; ok {
			db.userspace = us
			return nil
		}

		db.userspace = filepath.Join(rootpath, "userspace", uuid.New())

		userspaces[username] = db.userspace

		if err := db.WriteIndexer("userspace", userspaces); err != nil {
			return err
		}

		return nil
	})

	return db, err
}

// CacheRoot .
func (db *Metadata) CacheRoot() string {
	return filepath.Join(db.rootpath, "cache")
}

// Mount .
func (db *Metadata) Mount(src, target *Entry) error {

	gserrors.Require(target.Scheme == FSGSMake, "target must be rootfs node")

	indexername := path.Join(db.userspace, "mount")

	key := fmt.Sprintf("%s/%s", target.Host, target.Path)

	return db.Tx(func() error {

		var indexer map[string]string

		if err := db.ReadIndexer(indexername, &indexer); err != nil {
			return err
		}

		indexer[key] = src.String()

		if err := db.WriteIndexer(indexername, indexer); err != nil {
			return err
		}

		return nil
	})
}

// Dismount .
func (db *Metadata) Dismount(src, target *Entry) error {

	gserrors.Require(target.Scheme == FSGSMake, "target must be rootfs node")

	indexername := path.Join(db.userspace, "mount")

	key := fmt.Sprintf("%s/%s", target.Host, target.Path)

	return db.Tx(func() error {

		var indexer map[string]string

		if err := db.ReadIndexer(indexername, &indexer); err != nil {
			return err
		}

		delete(indexer, key)

		if err := db.WriteIndexer(indexername, indexer); err != nil {
			return err
		}

		return nil
	})
}

func (db *Metadata) queryMount(rootfs *rootfService, target *Entry) (entry *Entry, err error) {

	gserrors.Require(target.Scheme == FSGSMake, "target must be rootfs node")

	indexername := path.Join(db.userspace, "mount")

	key := fmt.Sprintf("%s/%s", target.Host, target.Path)

	err = db.Tx(func() error {

		var indexer map[string]string

		if err := db.ReadIndexer(indexername, &indexer); err != nil {
			return err
		}

		if src, ok := indexer[key]; ok {
			entry, err = rootfs.parseurl(src)
			return nil
		}

		return nil
	})

	return
}

// Tx start a transaction
func (db *Metadata) Tx(f func() error) error {
	return fs.FLock(db.flocker, f)
}

// ReadIndexer .
func (db *Metadata) ReadIndexer(name string, indexer interface{}) error {
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

// WriteIndexer .
func (db *Metadata) WriteIndexer(name string, indexer interface{}) error {

	indexerfile := filepath.Join(db.dbpath, name+".id")

	content, err := json.Marshal(indexer)

	if err != nil {
		return gserrors.Newf(err, "marshal %s indexer error", name)
	}

	err = ioutil.WriteFile(indexerfile, content, 0755)

	if err != nil {
		return gserrors.Newf(err, "write %s indexer error", name)
	}

	return nil
}
