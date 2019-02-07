package main

import (
	"github.com/Bitspark/go-funk"
	"github.com/Bitspark/slang/pkg/elem"
	"github.com/Bitspark/slang/pkg/storage"
	"github.com/google/uuid"
	"github.com/stoewer/go-strcase"
	"io/ioutil"
	"log"
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
}

type DocumentGenerator struct {
	libDir        string
	docOpDir      string
	opTmpl        *template.Template
	operatorInfos []OperatorInfo
}

func main() {
	libDir := "D:/go-workspace/src/slang-lib/slang"
	docDir := "D:/Bitspark/bitspark-www/html/pages/slang/docs/"
	tplDir := "D:/Bitspark/bitspark-www/templates/"

	dg := makeDocumentGenerator(libDir, docDir, tplDir)

	dg.init()
	dg.collect()
	dg.generate()
}

func makeDocumentGenerator(libDir string, docDir string, tmplDir string) DocumentGenerator {
	bytesOperator, err := ioutil.ReadFile(path.Join(tmplDir, "operator.html"))
	if err != nil {
		panic(err)
	}

	opTmpl, err := template.New("OperatorInfo").Delims("[[", "]]").Parse(string(bytesOperator))
	if err != nil {
		panic(err)
	}

	return DocumentGenerator{
		libDir:   libDir,
		docOpDir: path.Join(docDir, "operator"),
		opTmpl:   opTmpl,
	}
}

func (dg *DocumentGenerator) init() {
	err := os.RemoveAll(dg.docOpDir)
	if err != nil {
		panic(err)
	}

	os.MkdirAll(dg.docOpDir, os.ModeDir)
}

func (dg *DocumentGenerator) collect() {
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

	for _, id := range uuids {
		op, err := store.Load(id)
		if err != nil {
			panic(err)
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

		opSlug := strcase.KebabCase(op.Meta.Name)

		opTags := []OperatorTag{}
		
		opInfo := OperatorInfo{
			UUID:             id.String(),
			Name:             op.Meta.Name,
			Type:             opType,
			Icon:             op.Meta.Icon,
			Description:      op.Meta.Description,
			ShortDescription: op.Meta.ShortDescription,
			Slug:             opSlug,
			Tags:             opTags,
		}

		dg.operatorInfos = append(dg.operatorInfos, opInfo)
	}

	if len(dg.operatorInfos) == 0 {
		panic("No operators found")
	}

	log.Printf("Collected %d operators (%d elementaries, %d libraries)\n", len(dg.operatorInfos), elementaries, libraries)
}

func (dg *DocumentGenerator) generate() {
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
	}

	log.Printf("Generated %d files\n", generated)
}
