package vfs

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/gsdocker/gslogger"
)

func TestNew(t *testing.T) {

	defer gslogger.Join()

	_, err := New("./test", "vfs.test")

	if err != nil {
		t.Fatal(err)
	}
}
func TestURL(t *testing.T) {

	u, err := url.Parse("git://github.com/gsmake/gsmake?url=http://github.com/gsmake/gsmake.git")

	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(u.Query().Get("url"))
}
