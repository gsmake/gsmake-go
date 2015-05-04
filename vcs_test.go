package gsmake

import "testing"

func TestGit(t *testing.T) {
	cmd := NewGitCmd()

	err := cmd.Create(Properties{
		"repo": "https://github.com/gsdocker/gslogger.git",
		"dir":  ".repo/src/github.com/gsdocker/gslogger",
	})

	if err != nil {
		t.Fatal(err)
	}

	err = cmd.Update(Properties{
		"dir": ".repo/src/github.com/gsdocker/gslogger",
	})

	if err != nil {
		t.Fatal(err)
	}
}
