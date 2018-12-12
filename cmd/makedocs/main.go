package main

import (
	"fmt"
	"github.com/Bitspark/slang/pkg/api"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/elem"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

type operatorEntry struct {
	Name        string
	FQName      string
	Path        string
	Def         core.OperatorDef
	Type        string
	IconClass   string
	PackageHTML string
}

func main() {
	tplOperatorPath := "C:/Users/julia_000/Desktop/slang-docs/operator.html"
	tplIndexPath := "C:/Users/julia_000/Desktop/slang-docs/index.html"
	docDir := "C:/Users/julia_000/Desktop/slang-docs/docs/"
	docOperatorDir := "C:/Users/julia_000/Desktop/slang-docs/docs/operator/"

	err := os.RemoveAll(docDir)
	if err != nil {
		panic(err)
	}

	os.MkdirAll(docDir, os.ModeDir)
	os.MkdirAll(docOperatorDir, os.ModeDir)

	bytesOperator, err := ioutil.ReadFile(tplOperatorPath)
	if err != nil {
		panic(err)
	}
	tmplOperator, err := template.New("Operator").Delims("[[", "]]").Parse(string(bytesOperator))
	if err != nil {
		panic(err)
	}

	bytesIndex, err := ioutil.ReadFile(tplIndexPath)
	if err != nil {
		panic(err)
	}
	tmplIndex, err := template.New("Index").Delims("[[", "]]").Parse(string(bytesIndex))
	if err != nil {
		panic(err)
	}

	e := api.NewEnviron()
	fmt.Println(e)

	operators, err := e.ListOperatorNames()
	if err != nil {
		panic(err)
	}

	var libraryOperators []operatorEntry
	for _, fqname := range operators {
		if e.IsLocalOperator(fqname) {
			continue
		}

		opDefFilePath, _, err := e.GetFilePathWithFileEnding(strings.Replace(fqname, ".", string(filepath.Separator), -1), "")
		if err != nil {
			fmt.Println("elementary:", fqname)
			continue
		}

		def, err := e.ReadOperatorDef(opDefFilePath, []string{})
		if err != nil {
			continue
		}

		p := strings.Replace(fqname, ".", "/", -1) + ".html"
		entry := operatorEntry{getName(fqname), fqname, p, def, "library", "fa-box-open", buildPackageLine(fqname)}
		writeOperatorDocFile(docOperatorDir, tmplOperator, entry)
		libraryOperators = append(libraryOperators, entry)
	}

	var elementaryOperators []operatorEntry
	for _, fqname := range elem.GetBuiltinNames() {
		def, err := elem.GetOperatorDef(fqname)
		if err != nil {
			continue
		}

		p := strings.Replace(fqname, ".", "/", -1) + ".html"
		entry := operatorEntry{getName(fqname), fqname, p, def, "elementary", "fa-box", buildPackageLine(fqname)}
		writeOperatorDocFile(docOperatorDir, tmplOperator, entry)
		elementaryOperators = append(elementaryOperators, entry)
	}

	sort.SliceStable(libraryOperators, func(i, j int) bool {
		return strings.Compare(libraryOperators[i].Name, libraryOperators[j].Name) == -1
	})

	writeIndexDocFile(docDir, tmplIndex, buildIndex("", "", append(elementaryOperators, libraryOperators...)))
}

func getName(fqname string) string {
	fqnameSplit := strings.Split(fqname, ".")
	return fqnameSplit[len(fqnameSplit)-1]
}

func writeOperatorDocFile(docDir string, tmpl *template.Template, entry operatorEntry) {
	os.MkdirAll(path.Dir(docDir + entry.Path), os.ModeDir)

	file, err := os.Create(docDir + entry.Path)
	if err != nil {
		panic(err)
	}

	tmpl.Execute(file, entry)
	file.Close()
}

func writeIndexDocFile(docDir string, tmpl *template.Template, index string) {
	file, err := os.Create(docDir + "index.html")
	if err != nil {
		panic(err)
	}

	tmpl.Execute(file, struct {
		Index string
	}{index})
	file.Close()
}

func buildPackageLine(fqname string) string {
	fqnameSplit := strings.Split(fqname, ".")
	line := ""
	if len(fqnameSplit) > 1 {
		for _, p := range fqnameSplit[0 : len(fqnameSplit)-1] {
			line += "<a href=\"#\">" + p + "</a>"
			line += "."
		}
	}
	line += fqnameSplit[len(fqnameSplit)-1]
	return line
}

func buildIndex(pkg string, name string, ops []operatorEntry) string {
	index := ""
	if pkg != "" {
		index += "<a href=\"#\" class=\"toggle\"><i class=\"fas fa-minus-square\"></i></a> "
		index += name
	}

	index += "<ul>"
	subsHandled := make(map[string]bool)
	for _, op := range ops {
		if pkg != "" && !strings.HasPrefix(op.FQName, pkg+".") {
			continue
		}

		remainder := op.FQName
		if pkg != "" {
			remainder = remainder[len(pkg)+1:]
		}
		remainderSplit := strings.Split(remainder, ".")

		fmt.Println(remainderSplit)

		if len(remainderSplit) > 1 {
			if _, ok := subsHandled[remainderSplit[0]]; !ok {
				subsHandled[remainderSplit[0]] = true
				index += "<li class=\"package\">"
				if pkg == "" {
					index += buildIndex(remainderSplit[0], remainderSplit[0], ops)
				} else {
					index += buildIndex(pkg+"."+remainderSplit[0], remainderSplit[0], ops)
				}
				index += "</li>"
			}
		} else {
			index += "<li class=\"operator\">"
			index += "<a href=\"operator/" + op.Path + "\">"
			index += "<i class=\"fas " + op.IconClass + "\"></i> "
			index += remainder
			index += "</a>"
			index += "</li>"
		}
	}
	index += "</ul>"
	return index
}
