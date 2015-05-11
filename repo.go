package gsmake

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/gsdocker/gserrors"
)

// SearchCmd check command env
func SearchCmd(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", gserrors.Newf(
			err,
			"can't found command %s ",
			name)
	}

	return path, nil
}

// SCM source control manager
type SCM interface {
	fmt.Stringer
	Cmd() string
	Get(url string, reppath string, version string, targetpath string) error
}

// Repository gsmake package repository proxy
type Repository struct {
	settings *Settings       // settings
	sites    map[string]Site // register package host sites
	scm      map[string]SCM  // register scm
}

func loadRepository(settings *Settings) (*Repository, error) {
	repo := &Repository{
		settings: settings,
	}

	repo.sites = map[string]Site{
		"github.com": {
			SCM:     "git",
			URL:     "https://${root}.git",
			Package: `^(?P<root>github\.com/[A-Za-z0-9_.\-]+/[A-Za-z0-9_.\-]+)(/[A-Za-z0-9_.\-]+)*$`,
		},
		"gopkg.in": {
			SCM:     "git",
			URL:     "https://${root}",
			Package: `^(?P<root>gopkg\.in/[A-Za-z0-9_.\-])$`,
		},

		"bitbucket.org": {
			SCM:     "git",
			URL:     "https://${root}.git",
			Package: `^(?P<root>bitbucket\.org/[A-Za-z0-9_.\-]+/[A-Za-z0-9_.\-]+)(/[A-Za-z0-9_.\-]+)*$`,
		},
	}

	git, err := newGitSCM()

	if err != nil {
		return nil, err
	}

	repo.scm = map[string]SCM{
		git.Cmd(): git,
	}

	return repo, nil
}

// Get search gsmake package by package name and package version
func (repo *Repository) Get(name string, version string, targetpath string) error {

	// TODO: first check the package's scm

	prefix := name[:strings.IndexRune(name, '/')]

	if site, ok := repo.sites[prefix]; ok {
		scm, ok := repo.scm[site.SCM]

		if !ok {
			return gserrors.Newf(ErrImportPath, "not support remote site(%s) scm : %s", prefix, site.SCM)
		}

		matcher, err := regexp.Compile(site.Package)

		if err != nil {
			return gserrors.Newf(err, "compile site(%s) import path regexp error", prefix)
		}

		m := matcher.FindStringSubmatch(name)

		if m == nil {
			return gserrors.Newf(ErrImportPath, "invalid import path for vcs site(%s) : %s", prefix, name)
		}

		// Build map of named subexpression matches for expand.
		properties := Properties{}

		for i, name := range matcher.SubexpNames() {

			if name != "" && properties[name] == nil {
				properties[name] = m[i]
			}
		}

		return scm.Get(Expand(site.URL, properties), repo.settings.repoPath(name), version, targetpath)
	}

	return gserrors.Newf(nil, "invalid import path %s", name)
}
