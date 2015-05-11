package gsmake

import (
	"os"
	"testing"

	"github.com/gsdocker/gslogger"
	"github.com/gsdocker/gsos"
)

func init() {
	os.MkdirAll(".repo/repo/github.com/gsdocker", 0755)

	os.RemoveAll(".repo/repo/github.com/gsdocker/gsmake")

	os.Symlink(gsos.CurrentDir(), ".repo/repo/github.com/gsdocker/gsmake")
}

func TestLoad(t *testing.T) {
	defer gslogger.Join()
	compiler, err := Compile(".repo", "github.com/gsdocker/gsmake")

	if err != nil {
		t.Fatal(err)
	}

	compiler.Run("list")
}
