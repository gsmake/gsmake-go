package gsmake

import "testing"

func TestLoad(t *testing.T) {
	_, err := load(".repo", "./")

	if err != nil {
		t.Fatal(err)
	}
}

func TestCompile(t *testing.T) {
	compiler, err := compile(".repo", "./")

	if err != nil {
		t.Fatal(err)
	}

	compiler.Run("./", "list")
}
