package builder

import (
	"os"
	"path/filepath"
	"testing"
)

func init() {
	fullpath, _ := filepath.Abs("../")

	rootPath, _ := filepath.Abs("../.repo/src")

	linkpath := filepath.Join(rootPath, "github.com/gsdocker/gsmake/current")

	os.MkdirAll(linkpath, 0755)

	os.Remove(linkpath)

	os.Symlink(fullpath, linkpath)
}

func TestBuilder(t *testing.T) {
	builder, err := NewBuilder("../.repo")

	if err != nil {
		t.Fatal(err)
	}

	if err := builder.Prepare("../basic"); err != nil {
		t.Fatal(err)
	}

	if err := builder.Create(); err != nil {
		t.Fatal(err)
	}

	if err := builder.Run("publish"); err != nil {
		t.Fatal(err)
	}
}
