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
	Create(repo string, name string, version string) (string, error)
	Get(repo string, name string, version string, target string) error
	Update(repo string, name string) error
}

// Repository gsmake package repository proxy
type Repository struct {
	homepath string          // gsmake home path
	sites    map[string]Site // register package host sites
	scm      map[string]SCM  // register scm
}

func newRepository(homepath string) (*Repository, error) {
	repo := &Repository{
		homepath: homepath,
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

	git, err := newGitSCM(homepath)

	if err != nil {
		return nil, err
	}

	repo.scm = map[string]SCM{
		git.Cmd(): git,
	}

	return repo, nil
}

func (repo *Repository) calcURL(name string) (SCM, string, error) {
	prefix := name[:strings.IndexRune(name, '/')]

	if site, ok := repo.sites[prefix]; ok {
		scm, ok := repo.scm[site.SCM]

		if !ok {
			return nil, "", gserrors.Newf(ErrImportPath, "not support remote site(%s) scm : %s", prefix, site.SCM)
		}

		matcher, err := regexp.Compile(site.Package)

		if err != nil {
			return nil, "", gserrors.Newf(err, "compile site(%s) import path regexp error", prefix)
		}

		m := matcher.FindStringSubmatch(name)

		if m == nil {
			return nil, "", gserrors.Newf(ErrImportPath, "invalid import path for vcs site(%s) : %s", prefix, name)
		}

		// Build map of named subexpression matches for expand.
		properties := Properties{}

		for i, name := range matcher.SubexpNames() {

			if name != "" && properties[name] == nil {
				properties[name] = m[i]
			}
		}

		return scm, Expand(site.URL, properties), nil
	}

	return nil, "", gserrors.Newf(nil, "invalid import path %s", name)
}

// Update update cached package
func (repo *Repository) Update(name string) error {

	scm, url, err := repo.calcURL(name)

	if err != nil {
		return nil
	}

	return scm.Update(url, name)
}

// Create create cached package
func (repo *Repository) Create(name string, version string) (string, error) {

	scm, url, err := repo.calcURL(name)

	if err != nil {
		return "", nil
	}

	return scm.Create(url, name, version)
}

// Get search gsmake package by package name and package version
func (repo *Repository) Get(name string, version string, targetpath string) error {

	scm, url, err := repo.calcURL(name)

	if err != nil {
		return nil
	}

	return scm.Get(url, name, version, targetpath)
}
