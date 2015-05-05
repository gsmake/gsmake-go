package gsmake

import (
	"os"
	"path/filepath"
	"testing"
)

func init() {
	fullpath, _ := filepath.Abs("./")

	rootPath, _ := filepath.Abs(".repo/packages")

	linkpath := filepath.Join(rootPath, "github.com/gsdocker/gsmake/current")

	os.MkdirAll(linkpath, 0755)

	os.Remove(linkpath)

	os.Symlink(fullpath, linkpath)
}

func TestLoad(t *testing.T) {
	if _, err := Load(".repo", "./basic", false); err != nil {
		t.Fatal(err)
	}
}
func TestAOT(t *testing.T) {

	compiler, err := Compile(".repo", "./basic")

	if err != nil {
		t.Fatal(err)
	}

	if err := compiler.Run("-task"); err != nil {
		t.Fatal(err)
	}

	if err := compiler.Run("publish"); err != nil {
		t.Fatal(err)
	}
}
