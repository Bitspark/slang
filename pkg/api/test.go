package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"reflect"
	"github.com/Bitspark/slang/pkg/utils"
	"strings"
	"github.com/Bitspark/slang/pkg/core"
	"path/filepath"
)

type TestCaseDef struct {
	Name        string                   `json:"name" yaml:"name"`
	Description string                   `json:"description" yaml:"description"`
	Generics    map[string]*core.TypeDef `json:"generics" yaml:"generics"`
	Properties  map[string]interface{}   `json:"properties" yaml:"properties"`
	Data struct {
		In  []interface{} `json:"in" yaml:"in"`
		Out []interface{} `json:"out" yaml:"out"`
	}
}

type TestDef struct {
	OperatorFile string        `json:"operatorFile" yaml:"operatorFile"`
	Description  string        `json:"description" yaml:"description"`
	TestCases    []TestCaseDef `json:"testCases" yaml:"testCases"`
	valid        bool
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

	succs := 0
	fails := 0

	for i, tc := range test.TestCases {
		o, err := NewCustomEnviron("./").BuildAndCompileOperator(filepath.Join(filepath.Dir(testDataFilePath), test.OperatorFile), tc.Generics, tc.Properties)
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
	if len(t.OperatorFile) == 0 {
		return errors.New("no operator file given")
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
