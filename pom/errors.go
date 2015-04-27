package pom

import "errors"

// pom errors
var (
	ErrNotFound = errors.New("project not exists")
	ErrProject  = errors.New("project setting error")
)
