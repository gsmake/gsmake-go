package gsmake

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gsos"
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

	// Create create new repo
	Create(repo string, name string, version string) (string, error)
	// Create create update repo
	Update(repo string, name string, version string) (string, error)
	// Copy copy repo to target directory
	Copy(name string, version string, target string) error
	// Cache cache package as global repo package
	Cache(name string, version string, source string) error
	// RemoveCache remove cached package
	RemoveCache(name string, version string, source string) error
}

// Repository gsmake package repository proxy
type Repository struct {
	homepath  string              // gsmake home path
	sites     map[string]Site     // register package host sites
	scm       map[string]SCM      // register scm
	index     map[string][]string // repo index
	indexfile string              // index file path
}

func openRepository(homepath string) (*Repository, error) {
	repo := &Repository{
		homepath:  homepath,
		indexfile: filepath.Join(homepath, "repo", ".index"),
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

	return repo, repo.loadindex()
}

func (repo *Repository) loadindex() error {

	if gsos.IsExist(repo.indexfile) {
		content, err := ioutil.ReadFile(repo.indexfile)

		if err != nil {
			return gserrors.Newf(err, "read repo index file error")
		}

		err = json.Unmarshal(content, &repo.index)

		if err != nil {
			return gserrors.Newf(err, "read repo index file error")
		}
	} else {
		repo.index = make(map[string][]string)
	}

	return nil
}

func (repo *Repository) calcURL(name string) (SCM, string, error) {

	prefix := strings.SplitN(name, "/", 2)[0]

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
func (repo *Repository) Update(name string, version string) (string, error) {

	scm, url, err := repo.calcURL(name)

	if err != nil {
		return "", err
	}

	return scm.Update(url, name, version)
}

// Create create cached package
func (repo *Repository) Create(name string, version string) (string, error) {

	scm, url, err := repo.calcURL(name)

	if err != nil {
		return "", err
	}

	return repo.create(scm, url, name, version)

}

// Create create cached package
func (repo *Repository) create(scm SCM, url string, name string, version string) (string, error) {

	path, err := scm.Create(url, name, version)

	if err == nil {

		for _, v := range repo.index[name] {
			if v == version {
				return path, nil
			}
		}

		repo.index[name] = append(repo.index[name], version)

		content, err := json.Marshal(repo.index)

		if err != nil {
			return "", gserrors.Newf(err, "marshal repo index error")
		}

		return path, ioutil.WriteFile(repo.indexfile, content, 0644)

	}
	return path, err
}

// Cache cache pakage as global repo package
func (repo *Repository) Cache(name string, version string, source string) error {

	scm, _, err := repo.calcURL(name)

	if err != nil {
		return err
	}

	return scm.Cache(name, version, source)
}

// RemoveCache .
func (repo *Repository) RemoveCache(name string, version string, source string) error {

	scm, _, err := repo.calcURL(name)

	if err != nil {
		return err
	}

	return scm.RemoveCache(name, version, source)
}

// UpdateAll .
func (repo *Repository) UpdateAll() error {

	for name, versions := range repo.index {

		scm, url, err := repo.calcURL(name)

		if err != nil {
			return err
		}

		for _, version := range versions {
			_, err := scm.Update(url, name, version)

			if err != nil {
				return err
			}
		}

	}

	return nil
}

// Copy search gsmake package by package name and package version
func (repo *Repository) Copy(name string, version string, targetpath string) error {

	scm, url, err := repo.calcURL(name)

	if err != nil {
		return err
	}

	if _, err := repo.create(scm, url, name, version); err != nil {
		return err
	}

	return scm.Copy(name, version, targetpath)
}
