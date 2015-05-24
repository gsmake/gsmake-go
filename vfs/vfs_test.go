package vfs

import (
	"fmt"
	"testing"
)

func TestURL(t *testing.T) {
	url, err := Parse("sync://github.com/gsdocker/gsmake?version=master")

	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(url)

	url, err = Parse("sync:")

	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(url)
}

func TestVFS(t *testing.T) {
	service, err := NewVFService(".repo", "github.com/gsmake/gsmake.vfs.test")

	if err != nil {
		t.Fatal(err)
	}

	url, err := Parse("sync://github.com/gsdocker/gsmake?version=master")

	if err != nil {
		t.Fatal(err)
	}

	err = service.Mount("../", url)

	if err != nil {
		t.Fatal(err)
	}

	if err := service.Dismount("../", url); err != nil {
		t.Fatal(err)
	}
}
