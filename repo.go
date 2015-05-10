package gsmake

// SCM source control manager
type SCM interface {
}

// Repository gsmake package repository proxy
type Repository struct {
	settings Settings        // settings
	sites    map[string]Site // register package host sites
	scm      map[string]SCM  // register scm
}

func loadRepository(settings Settings) (*Repository, error) {
	return &Repository{
		sites: make(map[string]Site),
		scm:   make(map[string]SCM),
	}, nil
}

// Search search gsmake package by package name and package version
func (repo *Repository) Search(name string, version string) (string, error) {
	// TODO: first check the package's scm

	return "", nil
}
