package gsmake

import "errors"

// gsmake errors
var (
	ErrPackage    = errors.New("package error")
	ErrImportPath = errors.New("import path error")
	ErrVCSite     = errors.New("vcs site error")
	ErrTask       = errors.New("task error")
)
