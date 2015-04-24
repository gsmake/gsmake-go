package pom

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	path, _ := filepath.Abs("../")
	project, err := Parse(path)

	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("project: %v", project)
}
