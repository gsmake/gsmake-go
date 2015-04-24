package pom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/gsdocker/gserrors"
	"github.com/gsdocker/gslogger"
	"github.com/gsdocker/gsos"
)

// Parser the project ast parser
type Parser struct {
	gslogger.Log                        // Mixin log APIs
	projects        map[string]*Project // parsered project objects
	project         *Project            // root project
	linker          *Linker             // project linker
	parsingProjects []*Project          // parsing project
}

// Parse create new gsmake project parser
func Parse(rootPath string) (*Project, error) {

	linker, err := newLinker(rootPath)

	if err != nil {
		return nil, err
	}

	parser := &Parser{
		Log:      gslogger.Get("gsmake"),
		projects: make(map[string]*Project),
		linker:   linker,
	}

	parser.project, err = parser.parse(rootPath, nil)

	return parser.project, err
}

func (parser *Parser) parse(path string, importer *Project) (*Project, error) {

	if !gsos.IsDir(path) {
		return nil, gserrors.Newf(ErrLinker, "project not exist :\n\t%s", path)
	}

	gsmakefile := filepath.Join(path, ".gsmake.json")

	if !gsos.IsExist(gsmakefile) {
		return nil, gserrors.Newf(ErrLinker, "not found .gsmake file in project :\n\t%s", path)
	}

	content, err := ioutil.ReadFile(gsmakefile)

	if err != nil {
		return nil, err
	}

	project := new(Project)

	err = json.Unmarshal(content, &project)

	if err != nil {
		return nil, gserrors.Newf(err, "parse .gsmake.json file error\n\tfile :%s", gsmakefile)
	}

	// check must providing fields

	if project.Name == "" {
		return nil, gserrors.Newf(ErrParser, "must provide project name :\n\t%s", gsmakefile)
	}

	if project.Version == "" {
		return nil, gserrors.Newf(ErrParser, "must provide project version :\n\t%s", gsmakefile)
	}

	project.importer = importer
	project.path = path

	// check project cache
	if prev, ok := parser.projects[project.Name]; ok {
		if prev.Version == project.Version {
			return project, nil
		}

		return nil, gserrors.Newf(
			ErrParser,
			"duplicate import same project but different reversion\n%s\n%s",
			printImportPath(prev),
			printImportPath(project),
		)
	}

	// circle import check
	if err := parser.circleImportCheck(project); err != nil {
		return nil, err
	}

	parser.beginParse(project)
	defer gserrors.Assert(parser.endParse() == project, "inner check")

	for _, v := range project.Imports {
		if path, ok := parser.linker.link(project, v.Name, v.Version); ok {
			parser.parse(path, project)
		}
	}

	parser.projects[project.Name] = project

	return project, err
}

func (parser *Parser) circleImportCheck(project *Project) error {
	var stream bytes.Buffer

	for _, prev := range parser.parsingProjects {
		if prev.Name == project.Name || stream.Len() != 0 {
			stream.WriteString(fmt.Sprintf("\t%s import\n", prev.Name))
		}
	}

	if stream.Len() != 0 {
		return gserrors.Newf(ErrParser, "circular package import :\n%s\t%s", stream.String(), project.Name)
	}

	return nil
}

func printImportPath(project *Project) string {
	var stream bytes.Buffer

	for i := 0; project != nil; i++ {

		for j := 0; j < i; j++ {
			stream.WriteRune(' ')
		}

		stream.WriteString(project.String())
		stream.WriteString("->\n")
	}

	return stream.String()
}

func (parser *Parser) beginParse(project *Project) {
	parser.parsingProjects = append(parser.parsingProjects, project)
}

func (parser *Parser) endParse() *Project {
	project := parser.parsingProjects[len(parser.parsingProjects)-1]
	parser.parsingProjects = parser.parsingProjects[:len(parser.parsingProjects)-1]

	return project
}
