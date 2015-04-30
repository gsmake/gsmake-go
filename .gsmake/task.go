package plugin

import "github.com/gsdocker/gsmake"

// TaskResources .
func TaskResources(context *gsmake.Context) error {
	context.D("hello TaskResources")
	return nil
}

// TaskCompile .
func TaskCompile(context *gsmake.Context) error {
	context.D("hello TaskCompile")
	return nil
}

// TaskTest .
func TaskTest(context *gsmake.Context) error {
	context.D("hello test")
	return nil
}

// TaskPublish .
func TaskPublish(context *gsmake.Context) error {
	context.D("hello publish")
	return nil
}
