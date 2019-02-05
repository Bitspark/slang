package storage

import (
	"errors"
	"fmt"
	"github.com/Bitspark/go-funk"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/elem"
	"github.com/Bitspark/slang/pkg/utils"
	"github.com/google/uuid"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type TestLoader struct {
	// makes OperatorDef accessible by operator ID or operator Name
	dir     string
	storage map[string]core.OperatorDef
}

func NewTestLoader(dir string) *TestLoader {
	dir = filepath.Clean(dir)
	pathSep := string(filepath.Separator)
	if !strings.HasSuffix(dir, pathSep) {
		dir += pathSep
	}

	s := &TestLoader{dir, make(map[string]core.OperatorDef)}
	s.Reload()
	return s
}

func (tl *TestLoader) Reload() {
	tl.storage = make(map[string]core.OperatorDef)
	opDefList, err := readAllFiles(tl.dir)

	if err != nil {
		panic(err)
	}

	for _, opDef := range opDefList {
		opId := uuid.New()
		opDef.Id = opId.String()
		tl.storage[opDef.Id] = opDef
		tl.storage[opDef.Name] = opDef
	}

	// Replace instance operator names by ids
	for _, opDef := range opDefList {
		for _, childInsDef := range opDef.InstanceDefs {
			insOpId, err := uuid.Parse(childInsDef.Operator)

			if err == nil {
				childInsDef.Operator = insOpId.String()
				continue
			}

			insOpDef, ok := tl.storage[childInsDef.Operator]

			if ok {
				childInsDef.Operator = insOpDef.Id
				continue
			}

			if elemOpDef, err := elem.GetOperatorDef(childInsDef.Operator); err == nil {
				childInsDef.Operator = elemOpDef.Id
				continue
			}
		}
	}
}

func GetOperatorName(dir string, path string) string {
	relPath := strings.TrimSuffix(strings.TrimPrefix(path, dir), filepath.Ext(path))
	return strings.Replace(relPath, string(filepath.Separator), ".", -1)
}

func readAllFiles(dir string) ([]core.OperatorDef, error) {
	var opDefList []core.OperatorDef
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

		var opDef core.OperatorDef
		// Parse the file, just read it in
		if utils.IsYAML(path) {
			opDef, err = core.ParseYAMLOperatorDef(string(b))
		} else if utils.IsJSON(path) {
			opDef, err = core.ParseJSONOperatorDef(string(b))
		} else {
			err = errors.New("unsupported file ending")
		}
		if err != nil {
			return err
		}

		opDef.Name = GetOperatorName(dir, path)
		opDefList = append(opDefList, opDef)

		return nil
	})

	return opDefList, outerErr
}

func (tl *TestLoader) GetUUId(opName string) (uuid.UUID, bool) {
	opDef, ok := tl.storage[opName]
	id, _ := uuid.Parse(opDef.Id)
	return id, ok
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

func (tl *TestLoader) Load(opId uuid.UUID) (*core.OperatorDef, error) {
	if opDef, ok := tl.storage[opId.String()]; ok {
		return &opDef, nil
	}
	return nil, fmt.Errorf("unknown operator")
}
