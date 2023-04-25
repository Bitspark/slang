package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"text/template"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/elem"
	"github.com/Bitspark/slang/pkg/storage"
	"github.com/google/uuid"
	"github.com/stoewer/go-strcase"
	"github.com/thoas/go-funk"
)

type TagInfo struct {
	Tag           string
	Slug          string
	Count         int
	OperatorsJSON string

	operators []*OperatorInfo
}

type OperatorDefinition struct {
	ID   uuid.UUID
	Type string
	JSON string
}

type OperatorUsage struct {
	Count int
	Info  *OperatorInfo
}

type TestCase struct {
	Name        string
	Description string
	Data        []struct {
		In  string
		Out string
	}
}

type OperatorInfo struct {
	ID                  uuid.UUID
	Name                string
	Type                string
	Icon                string
	Description         string
	ShortDescription    string
	Slug                string
	Tags                []*TagInfo
	OperatorDefinitions []OperatorDefinition
	Tests               []TestCase

	OperatorContentCount int
	OperatorContentJSON  string
	OperatorsUsingCount  int
	OperatorsUsingJSON   string

	operatorContent map[uuid.UUID]*OperatorUsage
	operatorsUsing  map[uuid.UUID]*OperatorUsage

	operatorDefinition *core.Blueprint
}

type DocGenerator struct {
	libDir         string
	docOpDir       string
	docIndexPath   string
	docOpURL       *url.URL
	opTmpl         *template.Template
	operatorInfos  map[uuid.UUID]*OperatorInfo
	tagInfos       map[string]*TagInfo
	slugs          map[string]*OperatorInfo
	generatedInfos []*OperatorInfo
}

var clean bool
var genIdx bool
var saveUrls bool
var showHelp bool

var libDir string

var opTpl string
var opExt string
var opOutDir string

var idxTpl string
var idxOut string

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	docURL := "https://bitspark.de/slang/docs/"

	flag.BoolVar(&clean, "clean", false, "Clean folders before recreation")
	flag.BoolVar(&genIdx, "index", false, "Generate a single index file")
	flag.BoolVar(&saveUrls, "save-urls", false, "Save back doc urls into standard library")
	flag.BoolVar(&showHelp, "help", false, "Show this dialog")

	flag.StringVar(&libDir, "libdir", "./", "Location of the standard library")

	flag.StringVar(&idxOut, "index-target", "./", "Where to write the index to")
	flag.StringVar(&idxTpl, "index-template", "./", "Index template")

	flag.StringVar(&opTpl, "operator-template", "./", "Operator template")
	flag.StringVar(&opOutDir, "operator-output-dir", "./", "Where to write the operators to")
	flag.StringVar(&opExt, "operator-ext", "json", "What extension should the files have")
	flag.Parse()

	if showHelp {
		Usage()
		os.Exit(0)
	}

	dg := makeDocumentGenerator(libDir, idxOut, opTpl, opOutDir, docURL)
	if clean {
		dg.clean()
	}

	dg.collect(true)
	dg.contents()
	dg.usage()
	dg.generateOperatorDocs(opExt)

	if genIdx {
		bytesIndex, err := ioutil.ReadFile(idxTpl)
		if err != nil {
			panic(err)
		}
		indexTmpl, err := template.New("DocIndex").Delims("[[", "]]").Parse(string(bytesIndex))
		if err != nil {
			panic(err)
		}

		dg.generateIndex(*indexTmpl)
	}
	if saveUrls {
		dg.saveURLs()
	}

}

func makeDocumentGenerator(libDir string, idxOut string, opTpl string, opOutDir string, docURL string) DocGenerator {
	bytesOperator, err := ioutil.ReadFile(opTpl)
	if err != nil {
		panic(err)
	}
	opTmpl, err := template.New("DocOperatorInfo").Delims("[[", "]]").Parse(string(bytesOperator))
	if err != nil {
		panic(err)
	}

	docOpURL, _ := url.Parse(docURL)
	docOpURL.Path = path.Join(docOpURL.Path, "operator")

	return DocGenerator{
		libDir:        libDir,
		docOpDir:      opOutDir,
		docIndexPath:  idxOut,
		docOpURL:      docOpURL,
		opTmpl:        opTmpl,
		slugs:         make(map[string]*OperatorInfo),
		tagInfos:      make(map[string]*TagInfo),
		operatorInfos: make(map[uuid.UUID]*OperatorInfo),
	}
}

func (dg *DocGenerator) clean() {
	os.Remove(dg.docIndexPath)
	os.RemoveAll(dg.docOpDir)
}

func (dg *DocGenerator) collect(strict bool) {
	log.Println("Begin collecting")
	log.Printf("Library path: %s\n", dg.libDir)

	elementaryIDs := elem.GetBuiltinIds()

	store := storage.NewStorage().AddBackend(storage.NewReadOnlyFileSystem(dg.libDir))

	libraryIDs, err := store.List()
	if err != nil {
		panic(err)
	}

	var uuids []uuid.UUID

	for _, id := range elementaryIDs {
		uuids = append(uuids, id)
	}

	for _, id := range libraryIDs {
		uuids = append(uuids, id)
	}

	elementaries := 0
	libraries := 0
	tries := 0

	for _, id := range uuids {
		tries++

		blueprint, err := store.Load(id)
		if err != nil {
			log.Println(blueprint.Id, blueprint.Meta.Name, err)
			continue
		}

		if strict {
			if err := blueprint.Meta.Validate(); err != nil {
				log.Println(blueprint.Id, blueprint.Meta.Name, err)
				continue
			}
		}

		var opType string
		if funk.Contains(libraryIDs, id) {
			libraries++
			opType = "library"
		} else if funk.Contains(elementaryIDs, id) {
			elementaries++
			opType = "elementary"
		} else {
			panic("where did that uuid come from?!")
		}

		var opSlug string
		if blueprint.Meta.DocURL == "" {
			opSlug = dg.findSlug(blueprint, strcase.KebabCase(blueprint.Meta.Name))
		} else {
			u, err := url.Parse(blueprint.Meta.DocURL)
			if err != nil {
				panic(err)
			}
			opSlug = path.Base(u.Path)
		}

		opInfo := &OperatorInfo{}

		opTags := []*TagInfo{}
		for _, tag := range blueprint.Meta.Tags {
			kebabTag := strcase.KebabCase(tag)
			opTag, ok := dg.tagInfos[kebabTag]
			if !ok {
				opTag = &TagInfo{tag, kebabTag, 0, "", []*OperatorInfo{}}
				dg.tagInfos[kebabTag] = opTag
			}
			opTags = append(opTags, opTag)

			opTag.operators = append(opTag.operators, opInfo)
			opTag.Count++
		}

		opJSONDefs := make([]OperatorDefinition, 0)
		for _, jsonDef := range dumpDefinitions(id, store) {
			opJSONDefs = append(opJSONDefs, jsonDef)
		}

		opIcon := blueprint.Meta.Icon
		if opIcon == "" {
			opIcon = "box"
		}

		opTests := []TestCase{}

		for _, tc := range blueprint.TestCases {
			data := []struct {
				In  string
				Out string
			}{}

			for i := range tc.Data.In {
				buf := new(bytes.Buffer)

				json.NewEncoder(buf).Encode(tc.Data.In[i])
				buf.Truncate(buf.Len() - 1)
				inJSON := buf.String()

				buf.Reset()

				json.NewEncoder(buf).Encode(tc.Data.Out[i])
				buf.Truncate(buf.Len() - 1)
				outJSON := buf.String()

				data = append(data, struct {
					In  string
					Out string
				}{
					In:  inJSON,
					Out: outJSON,
				})
			}

			opTests = append(opTests, TestCase{
				Name:        tc.Name,
				Description: tc.Description,
				Data:        data,
			})
		}

		*opInfo = OperatorInfo{
			ID:                  id,
			Name:                blueprint.Meta.Name,
			Icon:                opIcon,
			Description:         blueprint.Meta.Description,
			ShortDescription:    blueprint.Meta.ShortDescription,
			Type:                opType,
			Slug:                opSlug,
			Tags:                opTags,
			Tests:               opTests,
			OperatorDefinitions: opJSONDefs,
			operatorDefinition:  blueprint,
			operatorContent:     make(map[uuid.UUID]*OperatorUsage),
			operatorsUsing:      make(map[uuid.UUID]*OperatorUsage),
		}

		dg.slugs[opSlug] = opInfo
		dg.operatorInfos[blueprint.Id] = opInfo
	}

	if len(dg.operatorInfos) == 0 {
		panic("No operators found")
	}

	fails := tries - len(dg.operatorInfos)

	log.Printf("Collected %d operators (%d elementaries, %d libraries)\n", len(dg.operatorInfos), elementaries, libraries)
	log.Printf("Failed to collect %d operators\n", fails)
}

func (dg *DocGenerator) contents() {
	for _, info := range dg.operatorInfos {
		for _, ins := range info.operatorDefinition.InstanceDefs {
			if usage, ok := info.operatorContent[ins.Operator]; ok {
				usage.Count++
			} else {
				info.operatorContent[ins.Operator] = &OperatorUsage{
					Count: 1,
					Info:  dg.operatorInfos[ins.Operator],
				}
			}
			info.OperatorContentCount++
		}
	}

	// Dump JSON
	for _, info := range dg.operatorInfos {
		buf := new(bytes.Buffer)
		json.NewEncoder(buf).Encode(info.operatorContent)
		buf.Truncate(buf.Len() - 1)
		info.OperatorContentJSON = buf.String()
	}
}

func (dg *DocGenerator) usage() {
	for _, info := range dg.operatorInfos {
		for _, ins := range info.operatorDefinition.InstanceDefs {
			insInfo, ok := dg.operatorInfos[ins.Operator]
			if !ok {
				continue
			}

			if usage, ok := insInfo.operatorsUsing[info.ID]; ok {
				usage.Count++
			} else {
				insInfo.operatorsUsing[info.ID] = &OperatorUsage{
					Count: 1,
					Info:  info,
				}
			}

			insInfo.OperatorsUsingCount++
		}
	}

	// Dump JSON
	for _, info := range dg.operatorInfos {
		buf := new(bytes.Buffer)
		json.NewEncoder(buf).Encode(info.operatorsUsing)
		buf.Truncate(buf.Len() - 1)
		info.OperatorsUsingJSON = buf.String()
	}
}

func (dg *DocGenerator) generateOperatorDocs(extension string) {
	log.Println("Begin generating operator docs")

	if len(dg.operatorInfos) == 0 {
		panic("No operators found")
	}

	//os.MkdirAll(dg.docOpDir, os.ModeDir)

	generated := 0

	for _, opInfo := range dg.operatorInfos {
		file, err := os.Create(path.Join(dg.docOpDir, opInfo.Slug+"."+extension))
		if err != nil {
			panic(err)
		}
		err = dg.opTmpl.Execute(file, opInfo)
		if err != nil {
			panic(err)
		}
		file.Close()

		generated++

		dg.generatedInfos = append(dg.generatedInfos, opInfo)
	}

	log.Printf("Generated %d operator doc files\n", generated)
}

func (dg *DocGenerator) prepareTags() {
	for _, tagInfo := range dg.tagInfos {
		// Remove JSON and tags to avoid recursion
		for i, op := range tagInfo.operators {
			opCpy := &OperatorInfo{}
			*opCpy = *op
			opCpy.OperatorContentJSON = ""
			opCpy.OperatorsUsingJSON = ""
			opCpy.OperatorDefinitions = nil
			tagInfo.operators[i] = opCpy
		}

		buf := new(bytes.Buffer)
		json.NewEncoder(buf).Encode(tagInfo.operators)
		buf.Truncate(buf.Len() - 1)
		tagInfo.OperatorsJSON = buf.String()
	}
}

func (dg *DocGenerator) generateIndex(indexTmpl template.Template) {
	log.Println("Begin generating doc index")

	if len(dg.tagInfos) == 0 {
		panic("No tags found")
	}
	dg.prepareTags()

	os.MkdirAll(path.Dir(dg.docIndexPath), os.ModeDir)

	file, err := os.Create(dg.docIndexPath)
	if err != nil {
		panic(err)
	}
	err = indexTmpl.Execute(file, struct {
		Total int
		Tags  map[string]*TagInfo
	}{len(dg.generatedInfos), dg.tagInfos})
	if err != nil {
		panic(err)
	}
	file.Close()

	log.Println("Generated doc index file")
}

func (dg *DocGenerator) saveURLs() {
	log.Println("Begin saving URLs")

	if len(dg.generatedInfos) == 0 {
		panic("No operators generated")
	}

	written := 0

	store := storage.NewStorage().AddBackend(storage.NewWritableFileSystem(dg.libDir))

	for _, opInfo := range dg.generatedInfos {
		if opInfo.Type != "library" {
			continue
		}

		blueprint := opInfo.operatorDefinition.Copy(false)

		opDocURL, _ := url.Parse(dg.docOpURL.String())
		opDocURL.Path = path.Join(opDocURL.Path, opInfo.Slug)
		opDocURLStr := opDocURL.String()

		//nolint:staticcheck
		if blueprint.Meta.DocURL == opDocURLStr {
			// continue
		}

		blueprint.Meta.DocURL = opDocURLStr

		_, err := store.Save(blueprint)
		if err != nil {
			panic(err)
		}

		written++
	}

	log.Printf("Updated %d URLs\n", written)
}

func (dg *DocGenerator) findSlug(blueprint *core.Blueprint, slug string) string {
	if info, ok := dg.slugs[slug]; !ok {
		return slug
	} else {
		otherTags := info.Tags
		additionalTags := []string{}

		for _, tag := range blueprint.Meta.Tags {
			if !funk.Contains(otherTags, tag) {
				additionalTags = append(additionalTags, tag)
			}
		}

		if len(additionalTags) == 0 {
			log.Panicf("cannot find alternative slug for")
		}

		slug += "-" + additionalTags[0]

		return dg.findSlug(blueprint, slug)
	}
}

func dumpDefinitions(id uuid.UUID, store *storage.Storage) map[uuid.UUID]OperatorDefinition {
	blueprint, err := store.Load(id)
	if err != nil {
		panic(err)
	}

	defs := make(map[uuid.UUID]OperatorDefinition)

	var opType string
	if blueprint.Elementary == uuid.Nil {
		opType = "library"
	} else {
		opType = "elementary"
	}

	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(blueprint)

	// Remove newline at the end
	buf.Truncate(buf.Len() - 1)

	defs[blueprint.Id] = OperatorDefinition{blueprint.Id, opType, buf.String()}

	for _, ins := range blueprint.InstanceDefs {
		subDefs := dumpDefinitions(ins.Operator, store)
		for id, def := range subDefs {
			if _, ok := defs[id]; !ok {
				defs[id] = def
			}
		}
	}

	return defs
}
