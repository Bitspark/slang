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

type TagInfo struct {
	Tag           string
	Slug          string
	Count         int
	OperatorsJSON string

	operators []*OperatorInfo
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

type TestCase struct {
	Name        string
	Description string
	Data        []struct {
		In  string
		Out string
	}
}

type OperatorInfo struct {
	UUID                string
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

	operatorContent map[string]*OperatorUsage
	operatorsUsing  map[string]*OperatorUsage

	operatorDefinition *core.OperatorDef
}

type DocGenerator struct {
	libDir         string
	docOpDir       string
	docTagDir      string
	docIndexPath   string
	docOpURL       *url.URL
	docTagURL      *url.URL
	opTmpl         *template.Template
	tagTmpl        *template.Template
	indexTmpl      *template.Template
	operatorInfos  map[string]*OperatorInfo
	tagInfos       map[string]*TagInfo
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
	dg.generateOperatorDocs()
	dg.prepareTags()
	// dg.generateTagDocs()
	dg.generateIndex()
	dg.saveURLs()
}

func makeDocumentGenerator(libDir string, docDir string, tmplDir string, docURL string) DocGenerator {
	bytesOperator, err := ioutil.ReadFile(path.Join(tmplDir, "operator.html"))
	if err != nil {
		panic(err)
	}
	opTmpl, err := template.New("DocOperatorInfo").Delims("[[", "]]").Parse(string(bytesOperator))
	if err != nil {
		panic(err)
	}

	//bytesTag, err := ioutil.ReadFile(path.Join(tmplDir, "tag.html"))
	//if err != nil {
	//	panic(err)
	//}
	//tagTmpl, err := template.New("DocTagInfo").Delims("[[", "]]").Parse(string(bytesTag))
	//if err != nil {
	//	panic(err)
	//}

	bytesIndex, err := ioutil.ReadFile(path.Join(tmplDir, "doc-index.html"))
	if err != nil {
		panic(err)
	}
	indexTmpl, err := template.New("DocIndex").Delims("[[", "]]").Parse(string(bytesIndex))
	if err != nil {
		panic(err)
	}

	docOpURL, _ := url.Parse(docURL)
	docOpURL.Path = path.Join(docOpURL.Path, "operator")
	docTagURL, _ := url.Parse(docURL)
	docTagURL.Path = path.Join(docTagURL.Path, "tag")

	return DocGenerator{
		libDir:        libDir,
		docOpDir:      path.Join(docDir, "operator"),
		docTagDir:     path.Join(docDir, "tag"),
		docIndexPath:  path.Join(docDir, "index.html"),
		docOpURL:      docOpURL,
		docTagURL:     docTagURL,
		opTmpl:        opTmpl,
		indexTmpl:     indexTmpl,
		slugs:         make(map[string]*OperatorInfo),
		tagInfos:      make(map[string]*TagInfo),
		operatorInfos: make(map[string]*OperatorInfo),
	}
}

func (dg *DocGenerator) init() {
	os.Remove(dg.docIndexPath)
	os.RemoveAll(dg.docOpDir)
	os.RemoveAll(dg.docTagDir)
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
	tries := 0

	for _, id := range uuids {
		tries++

		opDef, err := store.Load(id)
		if err != nil {
			log.Println(opDef.Id, opDef.Meta.Name, err)
			continue
		}

		if strict {
			if err := opDef.Meta.Validate(); err != nil {
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

		opInfo := &OperatorInfo{}

		opTags := []*TagInfo{}
		for _, tag := range opDef.Meta.Tags {
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

		opIcon := opDef.Meta.Icon
		if opIcon == "" {
			opIcon = "box"
		}

		opTests := []TestCase{}

		for _, tc := range opDef.TestCases {
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
			UUID:                id.String(),
			Name:                opDef.Meta.Name,
			Icon:                opIcon,
			Description:         opDef.Meta.Description,
			ShortDescription:    opDef.Meta.ShortDescription,
			Type:                opType,
			Slug:                opSlug,
			Tags:                opTags,
			Tests:               opTests,
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

			if usage, ok := insInfo.operatorsUsing[info.UUID]; ok {
				usage.Count++
			} else {
				insInfo.operatorsUsing[info.UUID] = &OperatorUsage{
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

func (dg *DocGenerator) generateOperatorDocs() {
	log.Println("Begin generating operator docs")

	if len(dg.operatorInfos) == 0 {
		panic("No operators found")
	}

	os.MkdirAll(dg.docOpDir, os.ModeDir)

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

func (dg *DocGenerator) generateTagDocs() {
	log.Println("Begin generating tag docs")

	if len(dg.tagInfos) == 0 {
		panic("No tags found")
	}

	os.MkdirAll(dg.docTagDir, os.ModeDir)

	generated := 0

	for _, tagInfo := range dg.tagInfos {
		file, err := os.Create(path.Join(dg.docTagDir, tagInfo.Slug+".html"))
		if err != nil {
			panic(err)
		}
		err = dg.tagTmpl.Execute(file, tagInfo)
		if err != nil {
			panic(err)
		}
		file.Close()

		generated++
	}

	log.Printf("Generated %d operator doc files\n", generated)
}

func (dg *DocGenerator) generateIndex() {
	log.Println("Begin generating doc index")

	if len(dg.tagInfos) == 0 {
		panic("No tags found")
	}

	os.MkdirAll(path.Dir(dg.docIndexPath), os.ModeDir)

	file, err := os.Create(dg.docIndexPath)
	if err != nil {
		panic(err)
	}
	err = dg.indexTmpl.Execute(file, struct {
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

	store := storage.NewStorage(storage.NewFileSystem(dg.libDir))

	for _, opInfo := range dg.generatedInfos {
		if opInfo.Type != "library" {
			continue
		}

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

func dumpDefinitions(id uuid.UUID, store *storage.Storage) map[string]OperatorDefinition {
	opDef, err := store.Load(id)
	if err != nil {
		panic(err)
	}

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
		opUuid, err := uuid.Parse(ins.Operator)
		if err != nil {
			panic(err)
		}
		subDefs := dumpDefinitions(opUuid, store)

		for id, def := range subDefs {
			if _, ok := defs[id]; !ok {
				defs[id] = def
			}
		}
	}

	return defs
}
