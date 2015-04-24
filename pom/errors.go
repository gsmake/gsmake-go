package pom

import "errors"

// errors
var (
	ErrLinker = errors.New("gsmake package link error")
	ErrParser = errors.New("gsmake package parse error")
)
