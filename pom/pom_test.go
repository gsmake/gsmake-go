//+

package pom

import "testing"

func TestReadProject(t *testing.T) {

	builder, err := NewBuilder("../")

	if err != nil {
		t.Fatal(err)
	}

	if err := builder.Compile(); err != nil {
		t.Fatal(err)
	}
}
