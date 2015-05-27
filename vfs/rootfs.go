package vfs

import (
	"fmt"
	"net/url"

	"github.com/gsdocker/gslogger"
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

// RootFS the gsmake vfs rootfs object
type RootFS interface {
	// Mixin Log APIs
	gslogger.Log
	//Mount mount src fs on to rootfs node
	Mount(src, target string) error
	//Dismount dismount src fs from rootfs node
	Dismount(target string) error
	// Update update a package
	Update(src string) error
	// Commit push modify to global cache or remote repo
	Commit(src string) error
	// Metadata get rootfs metadata
	Metadata() *Metadata
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
	Update(rootfs RootFS, src *Entry, target *Entry) error
	// Commit push modify to global cache or remote repo
	Commit(rootfs RootFS, src *Entry, target *Entry) error
	// Update all cached packages
	UpdateAll(rootfs RootFS) error
}
