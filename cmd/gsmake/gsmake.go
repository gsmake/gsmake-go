package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
	"github.com/gsdocker/gsos/fs"
	"github.com/gsdocker/gsos/uuid"
	"github.com/gsmake/gsmake"
	"github.com/gsmake/gsmake/vfs"
)

var helpmsg = `
gsmake is a build automation software for golang and others
Usage:
    go [flags] task
Use "gsmake list" list all task
`

// ImportVars .
type ImportVars struct {
	imports []gsmake.Import // import packages
}

func (vars *ImportVars) String() string {
	return "imports"
}

// Set .
func (vars *ImportVars) Set(val string) error {

	var ir gsmake.Import

	if err := json.Unmarshal([]byte(val), &ir); err != nil {
		return gserrors.Newf(err, "unmarshal import json error\n%s", val)
	}

	vars.imports = append(vars.imports, ir)

	return nil
}

var importVars ImportVars

var clearflag = flag.Bool("clear", false, "clear usrspace")
var verbflag = flag.Bool("v", false, "print more debug information")
var rootflag = flag.String("root", "", "the gsmake's root path")

var versionflag = flag.Bool("version", false, "print more debug information")

func init() {
	flag.Var(&importVars, "I", "import addition package")
}

func readconfig(log gslogger.Log) (string, string) {

	defer func() {
		if e := recover(); e != nil {
			log.E("%s", gserrors.Newf(e.(error), ""))
			gslogger.Join()
			os.Exit(1)
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

	if fs.Exists(".gsmake.json") {
		fullpath, err := filepath.Abs("./")

		if err != nil {
			log.E("get package full path error\n%s", err)
			gserrors.Panicf(err, "get package full path error")
		}

		packagedir = fullpath

	} else {

		dir := uuid.New()

		packagedir = filepath.Join(os.TempDir(), dir)

		if err := os.MkdirAll(packagedir, 0755); err != nil {
			gserrors.Panicf(err, "create %s error", packagedir)
		}

		pkg := &gsmake.Package{
			Name: gsmake.PacakgeAnonymous,
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

	return homepath, packagedir
}

func main() {

	currentdir := fs.Current()

	gslogger.Console(gsmake.Logfmt, gsmake.LogTimefmt)

	log := gslogger.Get("gsmake")

	flag.Parse()

	if *versionflag {
		fmt.Printf("Copyright (c) 2015 gsmake version %s \n", gsmake.VersionGSMake)
		return
	}

	if !*verbflag {
		gslogger.NewFlags(gslogger.ASSERT | gslogger.ERROR | gslogger.WARN | gslogger.INFO)
	}

	rootpath, targetpath := readconfig(log)

	rootfs, err := vfs.New(rootpath, targetpath)

	if err != nil {
		log.E("%s", err)
		gslogger.Join()
		os.Exit(1)
	}

	if strings.HasPrefix(targetpath, os.TempDir()) {
		defer func() {
			rootfs.Clear()

			if e := recover(); e != nil {
				log.E("%s", e)
				gslogger.Join()
				os.Exit(1)
			}

			gslogger.Join()
		}()
	} else {
		defer func() {
			if e := recover(); e != nil {
				log.E("%s", e)
				gslogger.Join()
				os.Exit(1)
			}

			gslogger.Join()
		}()
	}

	if *clearflag {
		if err := rootfs.Clear(); err != nil {
			panic(err)
		}
	}

	log.I("build gsmake runner ...")

	startime := time.Now()

	compiler, err := gsmake.Compile(rootfs, importVars.imports)

	if err != nil {
		gserrors.Panic(err)
	}

	log.I("build gsmake runner -- success %v", time.Now().Sub(startime))

	args := []string{}

	if *verbflag {
		args = append(args, "-v")
	}

	args = append(args, flag.Args()...)

	log.I("exec gsmake runner ...")

	startime = time.Now()

	if err := compiler.Run(currentdir, args...); err != nil {
		gserrors.Panic(err)
	}

	log.I("exec gsmake runner -- success %v", time.Now().Sub(startime))
}
