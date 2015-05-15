package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gsdocker/gserrors"
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

	currentdir := gsos.CurrentDir()

	log := gslogger.Get("gsmake")

	flag.Parse()

	if !*verbflag {
		gslogger.NewFlags(gslogger.ASSERT | gslogger.ERROR | gslogger.WARN | gslogger.INFO)
	}

	defer func() {
		if e := recover(); e != nil {
			log.E("recover :%s", e)
			gslogger.Join()
			os.Exit(1)
		} else {
			gslogger.Join()
		}
	}()

	homepath := os.Getenv(gsmake.EnvHome)

	if homepath == "" {
		homepath = *rootflag
	}

	if homepath == "" {
		gserrors.Panicf(nil, "expect -root flag or set %s env variable", gsmake.EnvHome)
	}

	var packagedir string

	if gsos.IsExist(".gsmake.json") {
		fullpath, err := filepath.Abs("./")

		if err != nil {
			log.E("get package full path error\n%s", err)
			gserrors.Panicf(err, "get package full path error")
		}

		packagedir = fullpath
	} else {
		dir := uuid.NewV1().String()

		packagedir = filepath.Join(os.TempDir(), dir)

		if err := os.MkdirAll(packagedir, 0755); err != nil {
			gserrors.Panicf(err, "create %s error", packagedir)
		}

		pkg := &gsmake.Package{
			Name: "github.com/gsdocker/gsmake.tmp",
			Import: []gsmake.Import{
				{Name: "github.com/gsdocker/gsmake"},
			},
		}

		content, err := json.Marshal(pkg)

		if err != nil {
			gserrors.Panicf(err, "create .gsmake.json error")
		}

		jsonfile := filepath.Join(packagedir, ".gsmake.json")

		err = ioutil.WriteFile(jsonfile, content, 0644)

		if err != nil {
			gserrors.Panicf(err, "create %s error", jsonfile)
		}
	}

	log.D("package path :%s", packagedir)
	log.D("gsmake root path :%s", homepath)

	compiler, err := gsmake.Compile(homepath, packagedir)

	if err != nil {
		panic(err)
	}

	if *verbflag {
		args := append([]string{"-v"}, flag.Args()...)
		compiler.Run(currentdir, args...)
	} else {
		compiler.Run(currentdir, flag.Args()...)
	}

	log.D("exist gsmake")
}
