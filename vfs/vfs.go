package vfs

import (
	"errors"
	"net/url"

	"github.com/gsdocker/gserrors"
)

// FileNode .
type FileNode int

// Builtin vfs nodes
const (
	NodeSync    FileNode = iota // sync vfs root
	NodeTask                    // task vfs root
	NodeRT                      // runtimes vfs root
	NodeTemp                    // temp directory
	NodeUnknown                 // unknown vfs node
)

// Errs
var (
	ErrNodeName = errors.New("unknown vfs node name")
)

// URL .
type URL struct {
	Root    FileNode // vfs root node
	Host    string   // host
	Path    string   // path string
	Version string   // filenode version string
}

var nodeNames = map[string]FileNode{
	"sync":    NodeSync,
	"task":    NodeTask,
	"rt":      NodeRT,
	"temp":    NodeTemp,
	"unknown": NodeUnknown,
}

var nodeIDs = func() map[FileNode]string {

	ret := make(map[FileNode]string)

	for k, v := range nodeNames {
		ret[v] = k
	}

	return ret
}()

// Parse .
func Parse(path string) (*URL, error) {
	u, err := url.Parse(path)

	if err != nil {
		return nil, err
	}

	node, ok := nodeNames[u.Scheme]

	if !ok {
		return nil, gserrors.Newf(ErrNodeName, "unknown scheme :%s", u.Scheme)
	}

	vfsurl := &URL{
		Root:    node,
		Host:    u.Host,
		Path:    u.Path,
		Version: "current",
	}

	if version := u.Query().Get("version"); version != "" {
		vfsurl.Version = version
	}

	return vfsurl, nil
}

func (vfsurl *URL) String() string {
	u := &url.URL{
		Host: vfsurl.Host,
		Path: vfsurl.Path,
	}

	scheme, ok := nodeIDs[vfsurl.Root]

	if !ok {
		gserrors.Newf(ErrNodeName, "unknown url root :%d", vfsurl.Root)
	}

	u.Scheme = scheme

	u.Query().Set("version", vfsurl.Version)

	return u.String()
}

// VFS gsmake root virtual file system
type VFS interface {

	// Mount mount native os directory into vfs
	Mount(native string, target *URL) error
	// Dismount dismount native directory from vfs
	Dismount(native string, target *URL) error
	// Commit commit fs change
	Commit(path *URL) error
	// Update update vfs directory
	Update(path *URL) error
	// Create create new file node
	Create(path *URL) error
	// Native convert vfs path to native file path
	Native(path *URL) (string, error)
}
