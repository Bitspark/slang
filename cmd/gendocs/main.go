package main

import (
	"bytes"
	"encoding/json"
	"github.com/Bitspark/go-funk"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/elem"
	"github.com/Bitspark/slang/pkg/storage"
	"github.com/google/uuid"
	"github.com/stoewer/go-strcase"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"text/template"
)

type OperatorTag struct {
	Tag  string
	Slug string
}

type OperatorDefinition struct {
	UUID string
	Type string
	JSON string
}

type OperatorUsage struct {
	Count int
	Info  *OperatorInfo
}

type OperatorInfo struct {
	UUID                string
	Name                string
	Type                string
	Icon                string
	Description         string
	ShortDescription    string
	Slug                string
	Tags                []OperatorTag
	OperatorDefinitions []OperatorDefinition

	OperatorContentJSON string
	OperatorsUsingJSON  string

	operatorContent map[string]*OperatorUsage
	operatorsUsing  map[string]*OperatorUsage

	operatorDefinition *core.OperatorDef
}

type DocGenerator struct {
	libDir         string
	docOpDir       string
	docOpURL       *url.URL
	opTmpl         *template.Template
	operatorInfos  map[string]*OperatorInfo
	slugs          map[string]*OperatorInfo
	generatedInfos []*OperatorInfo
}

func main() {
	libDir := "C:/Users/julia_000/Go/src/slang-lib/slang"
	docDir := "C:/Bitspark/bitspark-www/html/pages/slang/docs/"
	tplDir := "C:/Bitspark/bitspark-www/templates/"
	docURL := "https://bitspark.de/slang/docs/"

	dg := makeDocumentGenerator(libDir, docDir, tplDir, docURL)

	dg.init()
	dg.collect(true)
	dg.contents()
	dg.usage()
	dg.generate()
	dg.saveURLs()
}

func makeDocumentGenerator(libDir string, docDir string, tmplDir string, docURL string) DocGenerator {
	bytesOperator, err := ioutil.ReadFile(path.Join(tmplDir, "operator.html"))
	if err != nil {
		panic(err)
	}

	opTmpl, err := template.New("OperatorInfo").Delims("[[", "]]").Parse(string(bytesOperator))
	if err != nil {
		panic(err)
	}

	docOpURL, err := url.Parse(docURL)
	if err != nil {
		panic(err)
	}
	docOpURL.Path = path.Join(docOpURL.Path, "operator")

	return DocGenerator{
		libDir:        libDir,
		docOpDir:      path.Join(docDir, "operator"),
		docOpURL:      docOpURL,
		opTmpl:        opTmpl,
		slugs:         make(map[string]*OperatorInfo),
		operatorInfos: make(map[string]*OperatorInfo),
	}
}

func (dg *DocGenerator) init() {
	err := os.RemoveAll(dg.docOpDir)
	if err != nil {
		panic(err)
	}

	os.MkdirAll(dg.docOpDir, os.ModeDir)
}

func (dg *DocGenerator) collect(strict bool) {
	log.Println("Begin collecting")
	log.Printf("Library path: %s\n", dg.libDir)

	elementaryUUIDs := elem.GetBuiltinIds()

	store := storage.NewStorage(nil).AddLoader(storage.NewFileSystem(dg.libDir))

	libraryUUIDs, err := store.List()
	if err != nil {
		panic(err)
	}

	var uuids []uuid.UUID

	for _, id := range elementaryUUIDs {
		uuids = append(uuids, id)
	}

	for _, id := range libraryUUIDs {
		uuids = append(uuids, id)
	}

	elementaries := 0
	libraries := 0
	fails := 0

	for _, id := range uuids {
		opDef, err := store.Load(id)
		if err != nil {
			fails++
			log.Println(opDef.Id, opDef.Meta.Name, err)
			continue
		}

		if strict {
			if err := opDef.Meta.Validate(); err != nil {
				fails++
				log.Println(opDef.Id, opDef.Meta.Name, err)
				continue
			}
		}

		var opType string
		if funk.Contains(libraryUUIDs, id) {
			libraries++
			opType = "library"
		} else if funk.Contains(elementaryUUIDs, id) {
			elementaries++
			opType = "elementary"
		} else {
			panic("where did that uuid come from?!")
		}

		var opSlug string
		if opDef.Meta.DocURL == "" {
			opSlug = dg.findSlug(opDef, strcase.KebabCase(opDef.Meta.Name))
		} else {
			u, err := url.Parse(opDef.Meta.DocURL)
			if err != nil {
				panic(err)
			}
			opSlug = path.Base(u.Path)
		}

		opTags := []OperatorTag{}
		for _, tag := range opDef.Meta.Tags {
			opTags = append(opTags, OperatorTag{tag, strcase.KebabCase(tag)})
		}

		opJSONDefs := make([]OperatorDefinition, 0)
		for _, jsonDef := range dumpDefinitions(*opDef) {
			opJSONDefs = append(opJSONDefs, jsonDef)
		}

		opIcon := opDef.Meta.Icon
		if opIcon == "" {
			opIcon = "box"
		}

		opInfo := &OperatorInfo{
			UUID:                id.String(),
			Name:                opDef.Meta.Name,
			Icon:                opIcon,
			Description:         opDef.Meta.Description,
			ShortDescription:    opDef.Meta.ShortDescription,
			Type:                opType,
			Slug:                opSlug,
			Tags:                opTags,
			OperatorDefinitions: opJSONDefs,
			operatorDefinition:  opDef,
			operatorContent:     make(map[string]*OperatorUsage),
			operatorsUsing:      make(map[string]*OperatorUsage),
		}

		dg.slugs[opSlug] = opInfo
		dg.operatorInfos[opDef.Id] = opInfo
	}

	if len(dg.operatorInfos) == 0 {
		panic("No operators found")
	}

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

			if usage, ok := insInfo.operatorsUsing[ins.Operator]; ok {
				usage.Count++
			} else {
				insInfo.operatorsUsing[ins.Operator] = &OperatorUsage{
					Count: 1,
					Info:  insInfo,
				}
			}
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

func (dg *DocGenerator) generate() {
	log.Println("Begin generating")

	if len(dg.operatorInfos) == 0 {
		panic("No operators found")
	}

	generated := 0

	for _, opInfo := range dg.operatorInfos {
		file, err := os.Create(path.Join(dg.docOpDir, opInfo.Slug+".html"))
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

	log.Printf("Generated %d files\n", generated)
}

func (dg *DocGenerator) saveURLs() {
	log.Println("Begin saving URLs")

	if len(dg.generatedInfos) == 0 {
		panic("No operators generated")
	}

	written := 0

	store := storage.NewStorage(storage.NewFileSystem(dg.libDir))

	for _, opInfo := range dg.generatedInfos {
		opDef := opInfo.operatorDefinition.Copy(false)

		opDocURL, _ := url.Parse(dg.docOpURL.String())
		opDocURL.Path = path.Join(opDocURL.Path, opInfo.Slug)
		opDocURLStr := opDocURL.String()

		if opDef.Meta.DocURL == opDocURLStr {
			// continue
		}

		opDef.Meta.DocURL = opDocURLStr

		_, err := store.Store(opDef)
		if err != nil {
			panic(err)
		}

		written++
	}

	log.Printf("Updated %d URLs\n", written)
}

func (dg *DocGenerator) findSlug(opDef *core.OperatorDef, slug string) string {
	if info, ok := dg.slugs[slug]; !ok {
		return slug
	} else {
		otherTags := info.Tags
		additionalTags := []string{}

		for _, tag := range opDef.Meta.Tags {
			if !funk.Contains(otherTags, tag) {
				additionalTags = append(additionalTags, tag)
			}
		}

		if len(additionalTags) == 0 {
			log.Panicf("cannot find alternative slug for")
		}

		slug += "-" + additionalTags[0]

		return dg.findSlug(opDef, slug)
	}
}

func dumpDefinitions(opDef core.OperatorDef) map[string]OperatorDefinition {
	defs := make(map[string]OperatorDefinition)

	var opType string
	if opDef.Elementary == "" {
		opType = "library"
	} else {
		opType = "elementary"
	}

	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(opDef)

	// Remove newline at the end
	buf.Truncate(buf.Len() - 1)

	defs[opDef.Id] = OperatorDefinition{opDef.Id, opType, buf.String()}

	for _, ins := range opDef.InstanceDefs {
		subDefs := dumpDefinitions(ins.OperatorDef)

		for id, def := range subDefs {
			if _, ok := defs[id]; !ok {
				defs[id] = def
			}
		}
	}

	return defs
}
