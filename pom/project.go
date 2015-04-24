package pom

import "fmt"

// Project gsmake project node
type Project struct {
	Name     string              // Project universal name
	Version  string              // Project version string
	Imports  map[string]*Project // Project import other projects
	path     string              // Project package 's filesystem path
	parser   *Parser             // project's parser
	importer *Project            // project's first import project
}

func (project *Project) String() string {
	return fmt.Sprintf("%s:%s", project.Name, project.Version)
}
