package gsmake

import (
	"regexp"
	"strings"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
)

// VCSite .
type VCSite struct {
	VCS    string                 // vcs command type
	Repo   string                 // repo url
	Import string                 // import path regex pattern
	check  func(Properties) error // addition repo check
}

// Downloader project downloader
type Downloader struct {
	gslogger.Log                   // mixin Log APIs
	sites        map[string]VCSite // wellknown vcs sites
	vcs          map[string]VCSCmd // register vcs
}

// NewDownloader create downloader
func NewDownloader() *Downloader {

	downloader := &Downloader{
		Log: gslogger.Get("gsmake"),
		sites: map[string]VCSite{
			"github.com/": {
				VCS:    "git",
				Repo:   "https://${root}.git",
				Import: `^(?P<root>github\.com/[A-Za-z0-9_.\-]+/[A-Za-z0-9_.\-]+)(/[A-Za-z0-9_.\-]+)*$`,
			},
			"gopkg.in/": {
				VCS:    "git",
				Repo:   "https://${root}",
				Import: `^(?P<root>gopkg\.in/[A-Za-z0-9_.\-])$`,
			},

			"bitbucket.org/": {
				VCS:    "git",
				Repo:   "https://${root}.git",
				Import: `^(?P<root>bitbucket\.org/[A-Za-z0-9_.\-]+/[A-Za-z0-9_.\-]+)(/[A-Za-z0-9_.\-]+)*$`,
			},
		},

		vcs: map[string]VCSCmd{
			"git": NewGitCmd(),
		},
	}

	return downloader
}

// VCSite register vcs site
func (downloader *Downloader) VCSite(prefix string, site VCSite) {
	downloader.sites[prefix] = site
}

// Download .
func (downloader *Downloader) Download(importpath string, version string, dir string) error {

	for prefix, site := range downloader.sites {
		if strings.HasPrefix(importpath, prefix) {

			vcs, ok := downloader.vcs[site.VCS]

			if !ok {
				return gserrors.Newf(ErrVCSite, "unsupport vcs site(%s) : %s", prefix, site.VCS)
			}

			matcher, err := regexp.Compile(site.Import)

			if err != nil {
				return gserrors.Newf(err, "compile vcsite(%s) import path regexp error", prefix)
			}

			m := matcher.FindStringSubmatch(importpath)

			if m == nil {
				return gserrors.Newf(ErrImportPath, "invalid import path for vcs site(%s) : %s", prefix, importpath)
			}

			// Build map of named subexpression matches for expand.
			properties := Properties{}

			for i, name := range matcher.SubexpNames() {

				if name != "" && properties[name] == nil {
					properties[name] = m[i]
				}
			}

			properties["repo"] = Expand(site.Repo, properties)

			if site.check != nil {
				if err := site.check(properties); err != nil {
					return err
				}
			}

			properties["version"] = version

			properties["dir"] = dir

			return vcs.Create(properties)
		}
	}

	return gserrors.Newf(ErrImportPath, "unknown vcs site : %s", importpath)
}
