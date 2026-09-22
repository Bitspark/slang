package tests

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bitspark/slang/pkg/api"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/elem"
	"github.com/Bitspark/slang/pkg/storage"
	"github.com/Bitspark/slang/pkg/utils"
	"github.com/google/uuid"
	"github.com/thoas/go-funk"
)

func validateJSONOperatorDef(jsonDef string) (core.Blueprint, error) {
	def, _ := core.ParseJSONOperatorDef(jsonDef)
	return def, def.Validate()
}

type TestLoader struct {
	// makes Blueprint accessible by operator ID or operator Name
	dir     string
	storage map[string]core.Blueprint
}

func NewTestLoader(dir string) *TestLoader {
	dir = filepath.Clean(dir)
	pathSep := string(filepath.Separator)
	if !strings.HasSuffix(dir, pathSep) {
		dir += pathSep
	}

	s := &TestLoader{dir, make(map[string]core.Blueprint)}
	s.Reload()
	return s
}

func (tl *TestLoader) Reload() {
	tl.storage = make(map[string]core.Blueprint)
	blueprints, err := readAllFiles(tl.dir)

	if err != nil {
		panic(err)
	}

	for _, blueprint := range blueprints {
		tl.storage[blueprint.Id.String()] = blueprint
		tl.storage[blueprint.Meta.Name] = blueprint

		fmt.Println("-->", blueprint.Id.String(), blueprint.Meta.Name)
	}

	for _, blueprint := range blueprints {
		for _, childInsDef := range blueprint.InstanceDefs {
			insBlueprint, ok := tl.storage[childInsDef.Operator.String()]

			if ok {
				childInsDef.Operator = insBlueprint.Id
				continue
			}

			if elemBlueprint, err := elem.GetBlueprint(childInsDef.Operator); err == nil {
				childInsDef.Operator = elemBlueprint.Id
				continue
			}
		}
	}
}

func GetOperatorName(dir string, path string) string {
	relPath := filepath.Clean(strings.TrimSuffix(strings.TrimPrefix(path, dir), filepath.Ext(path)))
	return strings.Replace(relPath, string(filepath.Separator), ".", -1)
}

func readAllFiles(dir string) ([]core.Blueprint, error) {
	var blueprints []core.Blueprint
	outerErr := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() ||
			strings.HasPrefix(info.Name(), ".") ||
			!(utils.IsYAML(path) || utils.IsJSON(path)) {
			return nil
		}

		b, err := ioutil.ReadFile(path)
		if err != nil {
			return errors.New("could not read operator file " + path)
		}

		var blueprint core.Blueprint
		// Parse the file, just read it in
		if utils.IsYAML(path) {
			blueprint, err = core.ParseYAMLOperatorDef(string(b))
		} else if utils.IsJSON(path) {
			blueprint, err = core.ParseJSONOperatorDef(string(b))
		} else {
			err = errors.New("unsupported file ending")
		}
		if err != nil {
			return err
		}

		blueprint.Meta.Name = GetOperatorName(dir, path)
		blueprints = append(blueprints, blueprint)

		return nil
	})

	return blueprints, outerErr
}

func (tl *TestLoader) GetUUId(opName string) (uuid.UUID, bool) {
	blueprint, ok := tl.storage[opName]
	return blueprint.Id, ok
}

func (tl *TestLoader) Has(opId uuid.UUID) bool {
	_, ok := tl.storage[opId.String()]
	return ok
}

func (tl *TestLoader) List() ([]uuid.UUID, error) {
	var uuidList []uuid.UUID

	for _, idOrName := range funk.Keys(tl.storage).([]string) {
		if id, err := uuid.Parse(idOrName); err == nil {
			uuidList = append(uuidList, id)
		}
	}

	return uuidList, nil
}

func (tl *TestLoader) Load(opId uuid.UUID) (*core.Blueprint, error) {
	if blueprint, ok := tl.storage[opId.String()]; ok {
		return &blueprint, nil
	}
	return nil, fmt.Errorf("unknown operator")
}

type testEnv struct {
	dir string

	load *TestLoader
	stor *storage.Storage
}

func (t testEnv) getUUIDFromFile(opFile string) uuid.UUID {
	opName := GetOperatorName(t.dir, opFile)
	opId, _ := t.load.GetUUId(opName)
	return opId
}

func (t testEnv) RunTestBench(opFile string, writer io.Writer, failFast bool) (int, int, error) {
	tb := api.NewTestBench(t.stor)
	opId := t.getUUIDFromFile(opFile)
	return tb.Run(opId, writer, failFast)
}

func (t testEnv) CompileFile(opFile string, gens map[string]*core.TypeDef, props map[string]interface{}) (*core.Operator, error) {
	return api.BuildAndCompile(t.getUUIDFromFile(opFile), gens, props, *st)
}

const testdir string = "./"

var tl = NewTestLoader(testdir)
var st = storage.NewStorage().AddBackend(tl)
var Test = testEnv{testdir, tl, st}
