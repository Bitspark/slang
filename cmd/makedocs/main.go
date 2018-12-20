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
	"unicode"
)

type operatorEntry struct {
	Name             string
	DisplayName      string
	FQName           string
	Path             string
	Def              core.OperatorDef
	Type             string
	IconClass        string
	PackageList      string
	ShortDescription string
	Description      string
	Index            string
}

func main() {
	tplOperatorPath := "D:/Bitspark/bitspark-www/templates/operator.html"
	tplPackagePath := "D:/Bitspark/bitspark-www/templates/package.html"

	docDir := "D:/Bitspark/bitspark-www/html/pages/slang/docs/"

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

	var ops []struct {
		FQName string
		Def    core.OperatorDef
		Type   string
	}

	libs, err := e.ListOperatorNames()
	if err != nil {
		panic(err)
	}
	for _, fqname := range libs {
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

		ops = append(ops, struct {
			FQName string
			Def    core.OperatorDef
			Type   string
		}{fqname, def, "Library"})
	}

	elems := elem.GetBuiltinNames()
	for _, fqname := range elems {
		def, err := elem.GetOperatorDef(fqname)
		if err != nil {
			continue
		}

		ops = append(ops, struct {
			FQName string
			Def    core.OperatorDef
			Type   string
		}{fqname, def, "Elementary"})
	}

	var opEntries []operatorEntry
	for _, op := range ops {
		fqname := op.FQName
		def := op.Def

		p := strings.Replace(fqname, ".", "/", -1) + ".html"
		entry := operatorEntry{
			def.Name,
			def.DisplayName,
			fqname,
			p,
			def,
			op.Type,
			"fas fa-" + def.Icon,
			buildPackageList(fqname),
			def.ShortDescription,
			def.Description,
			"",
		}
		if entry.Name == "" {
			entry.Name = getName(entry.FQName)
			entry.DisplayName = getName(entry.FQName)
		}
		if def.Icon == "" {
			entry.IconClass = "fas fa-box-open"
		}
		opEntries = append(opEntries, entry)
	}

	for i := range opEntries {
		opEntries[i].Index = buildIndex(e, opEntries[i].FQName, "slang", opEntries)
	}

	for _, entry := range opEntries {
		writeOperatorDocFile(docDir, tmplOperator, entry)
	}

	sort.SliceStable(ops, func(i, j int) bool {
		return strings.Compare(ops[i].FQName, ops[j].FQName) == -1
	})

	buildPackageFiles(e, docDir, tmplPackage, "slang", "slang", opEntries)
}

func getPath(fqname string) string {
	fqsplit := strings.Split(fqname, ".")
	lastPart := fqsplit[len(fqsplit)-1]
	if unicode.IsUpper(rune(lastPart[0])) {
		pathsplit := fqsplit[0 : len(fqsplit)-1]
		p := strings.Join(pathsplit, ".")
		return p
	} else {
		return fqname
	}
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

func buildPackageFile(e *api.Environ, docDir string, tmpl *template.Template, pkg string, ops []operatorEntry) {
	file, err := os.Create(docDir + "index.html")
	if err != nil {
		panic(err)
	}

	pkgDef := getPackageDef(e, pkg)

	tmpl.Execute(file, struct {
		Name             string
		DisplayName      string
		Description      string
		ShortDescription string
		Index            string
	}{pkgDef.Name, pkgDef.DisplayName, pkgDef.Description, pkgDef.ShortDescription, buildIndex(e, pkg, "slang", ops)})
	file.Close()
}

func buildPackageList(fqname string) string {
	fqnameSplit := strings.Split(fqname, ".")
	pstr := ""
	line := ""
	if len(fqnameSplit) > 1 {
		for _, p := range fqnameSplit[0 : len(fqnameSplit)-1] {
			pstr += p + "/"

			line += "  - name: " + p + "\n"
			line += "    path: " + pstr + "\n"
		}
	}
	return line[:len(line)-1]
}

func buildPackageFiles(e *api.Environ, docDir string, tmpl *template.Template, pkg string, name string, ops []operatorEntry) {
	buildPackageFile(e, docDir+strings.Replace(pkg, ".", "/", -1)+"/", tmpl, pkg, ops)

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
				if pkg == "" {
					buildPackageFiles(e, docDir, tmpl, remainderSplit[0], remainderSplit[0], ops)
				} else {
					buildPackageFiles(e, docDir, tmpl, pkg+"."+remainderSplit[0], remainderSplit[0], ops)
				}
			}
		}
	}
}

func buildIndex(e *api.Environ, self, pkg string, ops []operatorEntry) string {
	pkgDef := getPackageDef(e, pkg)

	dpkg := getPath(self)

	class := ""
	icon := ""
	if strings.HasPrefix(dpkg, pkg) {
		class = "expanded"
		icon = "fa-minus-square"
	} else {
		class = "collapsed"
		icon = "fa-plus-square"
	}

	index := ""
	if pkg != "" {
		index += "<button class=\"toggle\"><i class=\"fas " + icon + "\"></i></button> "

		if pkg == self {
			index += "<span class=\"package-header selected\"> "
			index += pkgDef.DisplayName
			index += "</span>"
		} else {
			index += "<span class=\"package-header\"> "
			index += "<a href=\"{{root}}slang/docs/" + strings.Replace(pkg, ".", "/", -1) + "/\">"
			index += pkgDef.DisplayName
			index += "</a>"
			index += "</span>"
		}
	}

	index += "<ul class=\"" + class + "\">"
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
					index += buildIndex(e, self, remainderSplit[0], ops)
				} else {
					index += buildIndex(e, self, pkg+"."+remainderSplit[0], ops)
				}
				index += "</li>"
			}
		} else {
			if op.FQName == self {
				index += "<li class=\"operator\">"
				index += "<span class=\"operator-header selected\"> "
				index += "<i class=\"" + op.IconClass + "\"></i> "
				index += op.DisplayName
				index += "</span> "
				index += "</li>"
			} else {
				index += "<li class=\"operator\">"
				index += "<span class=\"operator-header\"> "
				index += "<a href=\"{{root}}slang/docs/" + op.Path + "\">"
				index += "<i class=\"" + op.IconClass + "\"></i> "
				index += op.DisplayName
				index += "</a>"
				index += "</span> "
				index += "</li>"
			}
		}
	}
	index += "</ul>"
	return index
}
