package main

import (
	"encoding/json"
	"github.com/Bitspark/slang/pkg/api"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/elem"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"unicode"
)

type OperatorYAML struct {
	Name string
	Type string
	YAML string
}

type OperatorEntry struct {
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
	OperatorYAMLs    []OperatorYAML
}

type Operator struct {
	Name      string `yaml:"name" json:"name"`
	IconClass string `yaml:"iconClass" json:"iconClass"`
	Path      string `yaml:"path" json:"path"`
	Opened    bool   `yaml:"opened" json:"opened"`

	Type string `yaml:"type" json:"type"`
}

type Package struct {
	Name      string `yaml:"name" json:"name"`
	IconClass string `yaml:"iconClass" json:"iconClass"`
	Path      string `yaml:"path" json:"path"`
	Expanded  bool   `yaml:"expanded" json:"expanded"`
	Opened    bool   `yaml:"opened" json:"opened"`

	SubPackages []*Package  `yaml:"subPackages" json:"subPackages"`
	Operators   []*Operator `yaml:"operators" json:"operators"`
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

	var opEntries []OperatorEntry
	for _, op := range ops {
		fqname := op.FQName
		def := op.Def

		operatorYAML := make(map[string]OperatorYAML)
		packOperatorIntoYAML(e, fqname, operatorYAML, "local")

		var operatorYAMLs []OperatorYAML
		for _, o := range operatorYAML {
			operatorYAMLs = append(operatorYAMLs, o)
		}

		p := strings.Replace(fqname, ".", "/", -1) + ".html"
		entry := OperatorEntry{
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
			operatorYAMLs,
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
		opEntries[i].Index = packageToString(buildIndex(e, opEntries[i].FQName, "slang", opEntries))
	}

	for _, entry := range opEntries {
		writeOperatorDocFile(docDir, tmplOperator, entry)
	}

	sort.SliceStable(ops, func(i, j int) bool {
		return strings.Compare(ops[i].FQName, ops[j].FQName) == -1
	})

	buildPackageFiles(e, docDir, tmplPackage, "slang", "slang", opEntries)
}

func packageToString(pkgPtr *Package) string {
	pkgYaml, err := json.Marshal(pkgPtr)
	if err != nil {
		panic(err)
	}
	return string(pkgYaml)
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

func writeOperatorDocFile(docDir string, tmpl *template.Template, entry OperatorEntry) {
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

func buildPackageFile(e *api.Environ, docDir string, tmpl *template.Template, pkg string, ops []OperatorEntry) {
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
	}{pkgDef.Name, pkgDef.DisplayName, pkgDef.Description, pkgDef.ShortDescription, packageToString(buildIndex(e, pkg, "slang", ops))})
	file.Close()
}

func buildPackageList(fqname string) string {
	type entry struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}

	var packageList []entry

	fqnameSplit := strings.Split(fqname, ".")
	pstr := ""
	if len(fqnameSplit) > 1 {
		for _, p := range fqnameSplit[0 : len(fqnameSplit)-1] {
			pstr += p + "/"
			packageList = append(packageList, entry{p, pstr})
		}
	}
	packageListJson, err := json.Marshal(&packageList)
	if err != nil {
		panic(err)
	}
	return string(packageListJson)
}

func buildPackageFiles(e *api.Environ, docDir string, tmpl *template.Template, pkg string, name string, ops []OperatorEntry) {
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

func buildIndex(e *api.Environ, self, pkg string, ops []OperatorEntry) *Package {
	pkgPtr := new(Package)
	pkgDef := getPackageDef(e, pkg)

	dpkg := getPath(self)
	pkgPtr.Expanded = strings.HasPrefix(dpkg, pkg)

	if pkg != "" {
		pkgPtr.Name = pkgDef.DisplayName
		pkgPtr.Opened = pkg == self
		pkgPtr.Path = strings.Replace(pkg, ".", "/", -1)
	}

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
				subPkgName := ""
				if pkg == "" {
					subPkgName = remainderSplit[0]
				} else {
					subPkgName = pkg + "." + remainderSplit[0]
				}
				subPkgPtr := buildIndex(e, self, subPkgName, ops)
				pkgPtr.SubPackages = append(pkgPtr.SubPackages, subPkgPtr)
			}
		} else {
			opPtr := new(Operator)
			opPtr.Name = op.DisplayName
			opPtr.Path = op.Path
			opPtr.Opened = op.FQName == self

			pkgPtr.Operators = append(pkgPtr.Operators, opPtr)
		}
	}

	return pkgPtr
}

func packOperatorIntoYAML(e *api.Environ, fqop string, read map[string]OperatorYAML, tp string) error {
	if _, ok := read[fqop]; ok {
		return nil
	}

	relPath := strings.Replace(fqop, ".", string(filepath.Separator), -1)
	absPath, _, err := e.GetFilePathWithFileEnding(relPath, "")
	if err != nil {
		return err
	}

	fileContents, err := ioutil.ReadFile(absPath)
	if err != nil {
		return err
	}
	read[fqop] = OperatorYAML{
		fqop,
		tp,
		string(fileContents),
	}

	def, err := e.ReadOperatorDef(absPath, nil)
	if err != nil {
		return err
	}

	var baseFqop string
	dotIdx := strings.LastIndex(fqop, ".")
	if dotIdx != -1 {
		baseFqop = fqop[:dotIdx+1]
	}
	for _, ins := range def.InstanceDefs {
		if elem.IsRegistered(ins.Operator) {
			def, err := elem.GetOperatorDef(ins.Operator)
			if err != nil {
				return err
			}
			defYAML, err := yaml.Marshal(def)
			if err != nil {
				return err
			}
			read[ins.Operator] = OperatorYAML{
				ins.Operator,
				"elementary",
				string(defYAML),
			}
		} else {
			tp := ""
			if strings.HasPrefix(ins.Operator, "slang.") {
				tp = "library"
			} else {
				tp = "local"
			}
			if !strings.HasPrefix(ins.Operator, ".") {
				packOperatorIntoYAML(e, ins.Operator, read, tp)
			} else {
				packOperatorIntoYAML(e, baseFqop+ins.Operator[1:], read, tp)
			}
		}
	}

	return nil
}
