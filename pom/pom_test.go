package test

import "testing"

func TestReadProject(t *testing.T) {
	_, err := NewProject("./")

	if err != nil {
		t.Fatal(err)
	}
}
