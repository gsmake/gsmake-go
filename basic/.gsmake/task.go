package plugin

import "github.com/gsdocker/gsmake"

// TaskGs2go .
func TaskGs2go(context *gsmake.Runner, args ...string) error {
	context.D("hello basic TaskGs2go")
	return nil
}

// TaskCompile .
func TaskCompile(context *gsmake.Runner, args ...string) error {
	context.D("hello basic TaskCompile")
	return nil
}
