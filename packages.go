package gsmake

import "github.com/gsmake/gsmake/property"

// Import the gsmake import instruction description
type Import struct {
	Name    string // import package name
	Version string // import package version
	Domain  string // runtimes import flag, default is AOT import
	SCM     string // the source control manager type
	URL     string // remote url
}

// Task package defined task description
type Task struct {
	Prev        string // depend task name
	Description string // task description
	Domain      string // scope belongs to
	Package     string `json:"-"` // package name which defined this task
}

// Package describe a gsmake package object
type Package struct {
	Name       string              // package name string
	Domain     string              // package usage scope
	Import     []Import            // package import field
	Task       map[string]*Task    // package defined task
	Properties property.Properties // properties
	Version    string              // package version
	Redirect   *Import             // package redirect instruction
	loadPath   []*Package          // package load path
}
