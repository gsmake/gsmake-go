package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gsdocker/gslogger"
	"github.com/gsdocker/gsmake"
	"github.com/gsdocker/gsos"
	"github.com/gsdocker/gsos/uuid"
)

var helpmsg = `
gsmake is a build automation software for golang and others

Usage:

    go [flags] task

Use "gsmake list" list all task

`

var verbflag = flag.Bool("v", false, "print more debug information")
var rootflag = flag.String("root", "", "the gsmake's root path")

func main() {

	var err error

	var packagedir string

	var rootpath string

	var compiler *gsmake.AOTCompiler

	log := gslogger.Get("gsmake")

	defer gslogger.Join()

	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println(helpmsg)
		goto Error
	}

	// if flag.NArg() > 1 {
	// 	args = flag.Args()[1:]
	// 	args = append(args, flag.Arg(0))
	// } else {
	// 	args = flag.Args()
	// }

	if !*verbflag {
		gslogger.NewFlags(gslogger.ASSERT | gslogger.ERROR | gslogger.WARN | gslogger.INFO)
	}

	if *rootflag == "" {
		rootpath = os.Getenv("GSMAKE_HOME")
		if rootpath == "" {
			log.E("env variable GSMAKE_HOME not defined or call gsmake with -root flag")
			goto Error
		}
	} else {
		rootpath = *rootflag
	}

	// already in package dir
	if gsos.IsExist(".gsmake.json") {
		fullpath, err := filepath.Abs("./")

		if err != nil {
			log.E("get package full path error\n%s", err)
			goto Error
		}

		packagedir = fullpath
	} else {

		dir := uuid.NewV1().String()

		packagedir = filepath.Join(os.TempDir(), dir)

		if err := os.MkdirAll(packagedir, 0755); err != nil {
			log.E("create %s error\n%s", packagedir, err)
			goto Error
		}

		pkg := &gsmake.Package{
			Name: "github.com/gsdocker/gsmake.tmp",
			Import: []gsmake.Import{
				{Name: "github.com/gsdocker/gsmake"},
			},
		}

		content, err := json.Marshal(pkg)

		if err != nil {
			log.E("create .gsmake.json error\n%s", err)
			goto Error
		}

		jsonfile := filepath.Join(packagedir, ".gsmake.json")

		err = ioutil.WriteFile(jsonfile, content, 0644)

		if err != nil {
			log.E("create %s error\n%s", jsonfile, err)
			goto Error
		}
	}

	log.D("package path :%s", packagedir)
	log.D("gsmake root path :%s", rootpath)

	compiler, err = gsmake.Compile(rootpath, packagedir)

	if err != nil {
		log.E("%s", err)
		goto Error
	}

	if *verbflag {
		args := append([]string{"-v"}, flag.Args()...)
		compiler.Run(args...)
	} else {
		compiler.Run(flag.Args()...)
	}

	return

Error:
	gslogger.Join()
	os.Exit(1)
}
