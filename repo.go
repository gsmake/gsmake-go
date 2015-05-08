package gsmake

// Repository gsmake package repository proxy
type Repository struct {
	settings Settings
}

func loadRepository(settings Settings) (*Repository, error) {
	return nil, nil
}

// Search search gsmake package by package name and package version
func (repo *Repository) Search(name string, version string) (string, error) {
	return "", nil
}
