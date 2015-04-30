package builder

// Downloader package downloader
type Downloader struct {
	root string // gsmake root path
}

// NewDownloader create new downloader
func NewDownloader(root string) *Downloader {
	return &Downloader{
		root: root,
	}
}

// Download download
func (downloader *Downloader) Download(name string, version string) error {
	return nil
}
