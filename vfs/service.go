package vfs

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
	"github.com/gsdocker/gsos/uuid"
	"github.com/gsmake/gsmake/vfs/nfs"
)

// VFService .
type VFService struct {
	gslogger.Log        // mixin loggers
	rootpath     string // vfs root's native fs path
	groupath     string // group name
	flock        string // file locker path
}

// NewVFService create new vfs service
func NewVFService(rootpath string, groupname string) (*VFService, error) {

	fullpath, err := filepath.Abs(rootpath)

	if err != nil {
		return nil, err
	}

	service := &VFService{
		Log:      gslogger.Get("vfs"),
		rootpath: fullpath,
		flock:    filepath.Join(fullpath, ".gsmake", "lock"),
	}

	service.I("starting vfs service ...")

	metadataDir := filepath.Join(fullpath, ".gsmake")

	if !nfs.Exists(metadataDir) {
		err := nfs.MkdirAll(metadataDir, 0755)

		if err != nil {
			return nil, gserrors.Newf(err, "create gsmake vfs root mount path error")
		}
	}

	service.D("vfs root path : %s", service.rootpath)

	// calc group rootpath
	err = nfs.FLock(service.flock, func() error {
		groupindexer := make(map[string]string)

		indexerfile := filepath.Join(fullpath, ".gsmake", "group")

		if nfs.Exists(indexerfile) {
			content, err := ioutil.ReadFile(indexerfile)
			if err != nil {
				return gserrors.Newf(err, "read vfs .group indexer file error")
			}

			err = json.Unmarshal(content, &groupindexer)

			if err != nil {
				return gserrors.Newf(err, "unmarshal vfs .group indexer file error")
			}
		}

		dir, ok := groupindexer[groupname]

		if !ok {
			dir = uuid.New()
			groupindexer[groupname] = dir
		}

		content, err := json.Marshal(groupindexer)

		if err != nil {
			return gserrors.Newf(err, "marshal vfs .group indexer file error")
		}

		err = ioutil.WriteFile(indexerfile, content, 0644)

		if err != nil {
			return gserrors.Newf(err, "write vfs .group indexer file error")
		}

		service.groupath = filepath.Join(fullpath, "groups", dir)

		service.D("group path : %s", service.groupath)

		return nil
	})

	if err != nil {
		return nil, err
	}

	service.I("start vfs service -- success")

	return service, nil
}

// Native convert vfs path to native file path
func (service *VFService) Native(path *URL) (string, error) {

	var nativepath string

	switch path.Root {
	case NodeSync:
		nativepath = filepath.Join(service.rootpath, "sync", path.Host, path.Path)
	case NodeTask:
		nativepath = filepath.Join(service.groupath, "task/src", path.Host, path.Path)
	case NodeRT:
		nativepath = filepath.Join(service.groupath, "runtimes/src", path.Host, path.Path)
	case NodeTemp:
		nativepath = filepath.Join(service.groupath, "temp", path.Host, path.Path)
	default:
		return "", gserrors.Newf(ErrNodeName, "not support node :%s", path.Root)
	}

	service.D("vfs node\n\tnode :%s\n\tnative fs :%s", path, nativepath)

	return nativepath, nil
}

// Mount mount native os directory into vfs{}
func (service *VFService) Mount(native string, url *URL) error {

	fullpath, err := filepath.Abs(native)

	if err != nil {
		return gserrors.Newf(err, "mount vfs node error : can't get src fullpath")
	}

	service.I("mount vfs node\n\tsrc :%s\n\ttarget :%s", fullpath, url)

	target, err := service.Native(url)

	if err != nil {
		return err
	}

	if nfs.Exists(target) {
		err := nfs.RemoveAll(target)

		if err != nil {
			return gserrors.Newf(err, "remove vfs node error")
		}
	}

	if err := nfs.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return gserrors.Newf(err, "mount vfs node error")
	}

	if err := nfs.Symlink(fullpath, target); err != nil {
		return gserrors.Newf(err, "mount vfs node error")
	}

	return nil
}

// Dismount dismount native directory from vfs{}
func (service *VFService) Dismount(native string, url *URL) error {

	src, err := filepath.Abs(native)

	if err != nil {
		return gserrors.Newf(err, "mount vfs node error : can't get src fullpath")
	}

	service.I("dismount vfs node\n\tsrc :%s\n\ttarget :%s", src, url)

	target, err := service.Native(url)

	if err != nil {
		return err
	}

	if !nfs.SameFile(src, target) {
		service.D("skipp dismount -- target dismatch src")
		return nil
	}

	err = nfs.RemoveAll(target)

	if err != nil {
		return gserrors.Newf(err, "dismount vfs node error")
	}

	return nil
}

// Commit commit fs change{}
func (service *VFService) Commit(path *URL) error {
	return nil
}

// Update update vfs directory{}
func (service *VFService) Update(path *URL) error {
	return nil
}

// Create create new file node{}
func (service *VFService) Create(path *URL) error {

	switch path.Root {
	case NodeSync:
	case NodeRT, NodeTask:
	case NodeTemp:
		return gserrors.Newf(ErrNodeName, "fnode[temp] not support create command")
	}

	return nil
}
