package main

import (
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
	Name             string
	DisplayName      string
	FQName           string
	Path             string
	Def              core.OperatorDef
	Type             string
	IconClass        string
	PackageHTML      string
	ShortDescription string
	Description      string
}

func main() {
	tplOperatorPath := "C:/Bitspark/bitspark-www/html/pages/slang/doc-templates/operator.html"
	tplPackagePath := "C:/Bitspark/bitspark-www/html/pages/slang/doc-templates/package.html"

	docDir := "C:/Bitspark/bitspark-www/html/pages/slang/docs/"

	err := os.RemoveAll(docDir + "slang")
	if err != nil {
		panic(err)
	}

	os.MkdirAll(docDir, os.ModeDir)

	bytesOperator, err := ioutil.ReadFile(tplOperatorPath)
	if err != nil {
		panic(err)
	}
	tmplOperator, err := template.New("Operator").Delims("[[", "]]").Parse(string(bytesOperator))
	if err != nil {
		panic(err)
	}

	bytesPackage, err := ioutil.ReadFile(tplPackagePath)
	if err != nil {
		panic(err)
	}
	tmplPackage, err := template.New("Index").Delims("[[", "]]").Parse(string(bytesPackage))
	if err != nil {
		panic(err)
	}

	e := api.NewEnviron()

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
			continue
		}

		def, err := e.ReadOperatorDef(opDefFilePath, []string{})
		if err != nil {
			continue
		}

		p := strings.Replace(fqname, ".", "/", -1) + ".html"
		entry := operatorEntry{
			def.Name,
			def.DisplayName,
			fqname,
			p,
			def,
			"library",
			"fas fa-" + def.Icon,
			buildPackageLine(fqname),
			def.ShortDescription,
			def.Description,
		}
		if entry.Name == "" {
			entry.Name = getName(entry.FQName)
			entry.DisplayName = getName(entry.FQName)
		}
		if def.Icon == "" {
			entry.IconClass = "fas fa-box-open"
		}
		writeOperatorDocFile(docDir, tmplOperator, entry)
		libraryOperators = append(libraryOperators, entry)
	}

	var elementaryOperators []operatorEntry
	for _, fqname := range elem.GetBuiltinNames() {
		def, err := elem.GetOperatorDef(fqname)
		if err != nil {
			continue
		}

		p := strings.Replace(fqname, ".", "/", -1) + ".html"
		entry := operatorEntry{
			def.Name,
			def.DisplayName,
			fqname,
			p,
			def,
			"elementary",
			"fas fa-" + def.Icon,
			buildPackageLine(fqname),
			def.ShortDescription,
			def.Description,
		}
		if entry.Name == "" {
			entry.Name = getName(entry.FQName)
			entry.DisplayName = getName(entry.FQName)
		}
		if def.Icon == "" {
			entry.IconClass = "fas fa-box"
		}
		writeOperatorDocFile(docDir, tmplOperator, entry)
		elementaryOperators = append(elementaryOperators, entry)
	}

	sort.SliceStable(libraryOperators, func(i, j int) bool {
		return strings.Compare(libraryOperators[i].Name, libraryOperators[j].Name) == -1
	})

	allOperators := append(elementaryOperators, libraryOperators...)

	// createPackagePages(docDir, tmplPackage)

	// makePackageFiles(docDir + "slang/", tmplPackage, buildIndex("", "", allOperators))

	buildPackageFiles(e, docDir, tmplPackage, "slang", "slang", allOperators)
}

func getName(fqname string) string {
	fqnameSplit := strings.Split(fqname, ".")
	return fqnameSplit[len(fqnameSplit)-1]
}

func writeOperatorDocFile(docDir string, tmpl *template.Template, entry operatorEntry) {
	os.MkdirAll(path.Dir(docDir+entry.Path), os.ModeDir)

	file, err := os.Create(docDir + entry.Path)
	if err != nil {
		panic(err)
	}

	tmpl.Execute(file, entry)
	file.Close()
}

func getPackageDef(e *api.Environ, pkg string) core.PackageDef {
	pkgDefPath, _, err := e.GetFilePathWithFileEnding(strings.Replace(pkg, ".", "/", -1)+"/_package", "")
	if err != nil {
		panic(err)
	}

	pkgDef, err := e.ReadPackageDef(pkgDefPath)
	if err != nil {
		panic(err)
	}

	return pkgDef
}

func createPackageFile(e *api.Environ, docDir string, tmpl *template.Template, pkg string, ops []operatorEntry) {
	file, err := os.Create(docDir + "index.html")
	if err != nil {
		panic(err)
	}

	pkgDef := getPackageDef(e, pkg)

	tmpl.Execute(file, struct {
		DisplayName      string
		ShortDescription string
		Index            string
	}{pkgDef.DisplayName, pkgDef.ShortDescription, buildIndex(e, pkg, ops)})
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

func buildPackageFiles(e *api.Environ, docDir string, tmpl *template.Template, pkg string, name string, ops []operatorEntry) {
	createPackageFile(e, docDir+strings.Replace(pkg, ".", "/", -1)+"/", tmpl, pkg, ops)

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

		if len(remainderSplit) > 1 {
			if _, ok := subsHandled[remainderSplit[0]]; !ok {
				subsHandled[remainderSplit[0]] = true
			} else {
				if pkg == "" {
					buildPackageFiles(e, docDir, tmpl, remainderSplit[0], remainderSplit[0], ops)
				} else {
					buildPackageFiles(e, docDir, tmpl, pkg+"."+remainderSplit[0], remainderSplit[0], ops)
				}
			}
		}
	}
}

func buildIndex(e *api.Environ, pkg string, ops []operatorEntry) string {
	pkgDef := getPackageDef(e, pkg)

	index := ""
	if pkg != "" {
		index += "<button class=\"toggle\"><i class=\"fas fa-minus-square\"></i></button> "
		index += "<a href=\"{{root}}slang/docs/" + strings.Replace(pkg, ".", "/", -1) + "/\">"
		index += pkgDef.DisplayName
		index += "</a>"
	}

	index += "<ul class=\"expanded\">"
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

		if len(remainderSplit) > 1 {
			if _, ok := subsHandled[remainderSplit[0]]; !ok {
				subsHandled[remainderSplit[0]] = true
				index += "<li class=\"package\">"
				if pkg == "" {
					index += buildIndex(e, remainderSplit[0], ops)
				} else {
					index += buildIndex(e, pkg+"."+remainderSplit[0], ops)
				}
				index += "</li>"
			}
		} else {
			index += "<li class=\"operator\">"
			index += "<a href=\"{{root}}slang/docs/" + op.Path + "\">"
			index += "<i class=\"" + op.IconClass + "\"></i> "
			index += op.DisplayName
			index += "</a>"
			index += "</li>"
		}
	}
	index += "</ul>"
	return index
}
