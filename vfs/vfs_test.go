package vfs

import (
	"fmt"
	"testing"
)

func TestGitMount(t *testing.T) {
	rootfs, err := New("./test", "vfs.test")

	if err != nil {
		t.Fatal(err)
	}

	err = rootfs.Mount(
		"git://github.com/gsdocker/gserrors?remote=http://github.com/gsdocker/gserrors.git&version=v1.2",
		"gsmake://github.com/gsdocker/gserrors?domain=task&version=master",
	)

	if err != nil {
		t.Fatal(err)
	}

	err = rootfs.Update("gsmake://github.com/gsdocker/gserrors?domain=task", false)

	if err != nil {
		t.Fatal(err)
	}

	_, _, err = rootfs.Open("gsmake://github.com/gsdocker/gserrors?domain=task")

	if err != nil {
		t.Fatal(err)
	}

	_, _, err = rootfs.Open("gsmake://github.com/gsdocker/gserrors?domain=rc")

	if !NotFound(err) {
		t.Fatal("open not exists file error")
	}

	entries, err := rootfs.List()

	if err != nil {
		t.Fatal(err)
	}

	for k, v := range entries {
		fmt.Printf("%s -> %s\n", k, v)
	}

	err = rootfs.Clear()

	if err != nil {
		t.Fatal(err)
	}
}
