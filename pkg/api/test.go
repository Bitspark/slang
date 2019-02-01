package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Bitspark/go-funk"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/elem"
	"github.com/Bitspark/slang/pkg/utils"
	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type TestCaseDef struct {
	Name        string                   `json:"name" yaml:"name"`
	Description string                   `json:"description" yaml:"description"`
	Generics    map[string]*core.TypeDef `json:"generics" yaml:"generics"`
	Properties  map[string]interface{}   `json:"properties" yaml:"properties"`
	Data        struct {
		In  []interface{} `json:"in" yaml:"in"`
		Out []interface{} `json:"out" yaml:"out"`
	}
}

type TestDef struct {
	Operator    string        `json:"operator" yaml:"operator"`
	Description string        `json:"description" yaml:"description"`
	TestCases   []TestCaseDef `json:"testCases" yaml:"testCases"`
	valid       bool
}

type TestStorage struct {
	// makes OperatorDef accessible by operator ID or operator Name
	storage map[string]core.OperatorDef
}

func NewTestStorage(dir string) *TestStorage {
	dir = filepath.Clean(dir)
	pathSep := string(filepath.Separator)
	if !strings.HasSuffix(dir, pathSep) {
		dir += pathSep
	}
	storage := make(map[string]core.OperatorDef)
	opDefList, err := readAllFiles(dir)

	if err != nil {
		panic(err)
	}

	for _, opDef := range opDefList {
		opId := uuid.New()
		opDef.Id = opId.String()
		storage[opDef.Id] = opDef
		storage[opDef.Name] = opDef
	}

	for _, opDef := range opDefList {
		for _, childInsDef := range opDef.InstanceDefs {
			var childOpId string
			if childOpDef, ok := storage[childInsDef.Operator]; ok {
				childOpId = childOpDef.Id
			} else if elemOpDef, err := elem.GetOperatorDef(childInsDef.Operator); err == nil {
				childOpId = elemOpDef.Id
			}

			childInsDef.Operator = childOpId
		}
	}

	return &TestStorage{storage}
}

func readAllFiles(dir string) ([]core.OperatorDef, error) {
	var opDefList []core.OperatorDef
	outerErr := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() ||
			strings.HasPrefix(info.Name(), ".") ||
			!(strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".json")) {
			return nil
		}

		b, err := ioutil.ReadFile(path)
		if err != nil {
			return errors.New("could not read operator file " + path)
		}

		var opDef core.OperatorDef
		// Parse the file, just read it in
		if utils.IsYAML(path) {
			opDef, err = ParseYAMLOperatorDef(string(b))
		} else if utils.IsJSON(path) {
			opDef, err = ParseJSONOperatorDef(string(b))
		} else {
			err = errors.New("unsupported file ending")
		}
		if err != nil {
			return err
		}

		relPath := strings.TrimSuffix(strings.TrimPrefix(path, dir), filepath.Ext(path))
		opDef.Name = strings.Replace(relPath, string(filepath.Separator), ".", -1)
		opDefList = append(opDefList, opDef)

		return nil
	})

	return opDefList, outerErr
}

func (s *TestStorage) IsLibrary(opId uuid.UUID) bool {
	_, ok := s.storage[opId.String()]
	return !ok
}

func (s *TestStorage) List() ([]uuid.UUID, error) {
	var uuidList []uuid.UUID

	for _, idOrName := range funk.Keys(s.storage).([]string) {
		if id, err := uuid.Parse(idOrName); err == nil {
			uuidList = append(uuidList, id)
		}
	}

	return uuidList, nil
}

func (s *TestStorage) Load(opId uuid.UUID) (*core.OperatorDef, error) {

	if opDef, ok := s.storage[opId.String()]; ok {
		return &opDef, nil
	}

	return nil, fmt.Errorf("unknown operator for given id: %s", opId.String())
}

func (s *TestStorage) Store(opDef core.OperatorDef) (uuid.UUID, error) {
	panic("Not implemted")
	return uuid.UUID{}, nil
}

// TestOperator reads a file with test data and its corresponding operator and performs the tests.
// It returns the number of failed and succeeded tests and and error in case something went wrong.
// Test failures do not lead to an error. Test failures are printed to the writer.
func TestOperator(testDataFilePath string, writer io.Writer, failFast bool) (int, int, error) {
	b, err := ioutil.ReadFile(testDataFilePath)

	if err != nil {
		return 0, 0, err
	}

	test := TestDef{}
	if strings.HasSuffix(testDataFilePath, ".yaml") || strings.HasSuffix(testDataFilePath, ".yml") {
		err = yaml.Unmarshal(b, &test)
	} else if strings.HasSuffix(testDataFilePath, ".json") {
		err = json.Unmarshal(b, &test)
	} else {
		err = errors.New("unsupported file ending")
	}
	if err != nil {
		return 0, 0, err
	}

	if !test.Valid() {
		err := test.Validate()
		if err != nil {
			return 0, 0, err
		}
	}

	opId, _ := uuid.Parse(test.Operator)

	succs := 0
	fails := 0

	st := NewTestStorage("./")

	for i, tc := range test.TestCases {
		opDef, err := st.Load(opId)

		if err != nil {
			return 0, 0, err
		}

		o, err := BuildAndCompile(*opDef, tc.Generics, tc.Properties)
		if err != nil {
			return 0, 0, err
		}

		fmt.Fprintf(writer, "Test case %3d/%3d: %s (operators: %d, size: %d)\n", i+1, len(test.TestCases), tc.Name, len(o.Children()), len(tc.Data.In))

		if err := o.CorrectlyCompiled(); err != nil {
			return 0, 0, err
		}
		o.Main().Out().Bufferize()
		o.Start()

		success := true

		for j := range tc.Data.In {
			in := tc.Data.In[j]
			expected := utils.CleanValue(tc.Data.Out[j])

			o.Main().In().Push(utils.CleanValue(in))
			actual := o.Main().Out().Pull()

			if !testEqual(expected, actual) {
				fmt.Fprintf(writer, "  expected: %#v (%T)\n", expected, expected)
				fmt.Fprintf(writer, "  actual:   %#v (%T)\n", actual, actual)

				success = false

				if failFast {
					o.Stop()
					return succs, fails + 1, nil
				}
			}
		}

		if success {
			fmt.Fprintln(writer, "  success")
			succs++
		} else {
			fails++
		}

		o.Stop()
	}

	return succs, fails, nil
}

func (t TestDef) Validate() error {
	if t.Operator == "" {
		return fmt.Errorf(`no operator id given`)
	}

	if _, err := uuid.Parse(t.Operator); err != nil {
		return fmt.Errorf(`id is not a valid UUID v4: "%s" --> "%s"`, t.Operator, err)
	}

	for _, tc := range t.TestCases {
		if len(tc.Data.In) != len(tc.Data.Out) {
			return fmt.Errorf(`data count unequal in test case "%s"`, tc.Name)
		}
	}

	t.valid = true
	return nil
}

func (t TestDef) Valid() bool {
	return t.valid
}

func testEqual(a, b interface{}) bool {
	as, aok := a.([]interface{})
	bs, bok := b.([]interface{})

	if aok && bok {
		if len(as) != len(bs) {
			return false
		}

		for i, ai := range as {
			bi := bs[i]
			if !testEqual(ai, bi) {
				return false
			}
		}

		return true
	}

	am, aok := a.(map[string]interface{})
	bm, bok := b.(map[string]interface{})

	if aok && bok {
		if len(am) != len(bm) {
			return false
		}

		for k, ai := range am {
			if bi, ok := bm[k]; ok {
				if !testEqual(ai, bi) {
					return false
				}
			} else {
				return false
			}
		}

		return true
	}

	if ai, ok := a.(int); ok {
		a = float64(ai)
	}
	if bi, ok := b.(int); ok {
		b = float64(bi)
	}
	return reflect.DeepEqual(a, b)
}
