package gsmake

import (
	"strings"

	"github.com/gsdocker/gserrors"
)

// SCM source control manager
type SCM interface {
	Get(settings *Settings, name string, version string) (string, error)
}

// Repository gsmake package repository proxy
type Repository struct {
	settings *Settings       // settings
	sites    map[string]Site // register package host sites
	scm      map[string]SCM  // register scm
}

func loadRepository(settings *Settings) (*Repository, error) {
	return &Repository{
		settings: settings,
		sites:    make(map[string]Site),
		scm:      make(map[string]SCM),
	}, nil
}

// Search search gsmake package by package name and package version
func (repo *Repository) Search(name string, version string) (string, error) {
	// TODO: first check the package's scm

	if site, ok := repo.sites[name[:strings.IndexRune(name, '/')]]; ok {
		scm := repo.scm[site.SCM]

		path, err := scm.Get(repo.settings, name, version)

		return path, err
	}

	return "", gserrors.Newf(nil, "invalid import path %s", name)
}
