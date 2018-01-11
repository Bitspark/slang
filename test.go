package slang

import (
	"io"
	"io/ioutil"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"reflect"
	"strings"
	"gopkg.in/yaml.v2"
	"slang/utils"
)

type TestCaseDef struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
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

	o, err := BuildOperator(path.Join(path.Dir(testDataFilePath), test.OperatorFile), false)

	if err != nil {
		return 0, 0, err
	}

	fmt.Fprintf(writer, "%s parsed successfully\n", test.OperatorFile)

	compiled := o.Compile()
	fmt.Fprintln(writer, "Operator compiled")
	fmt.Fprintf(writer, "Operators simplified:  %3d\n", compiled)
	fmt.Fprintf(writer, "Total basic operators: %3d\n", len(o.Children()))

	o.Out().Bufferize()

	o.Start()
	defer o.Stop()
	fmt.Fprintln(writer, "Operator started")
	fmt.Fprintln(writer)

	succs := 0
	fails := 0

	fmt.Fprintln(writer, "BEGIN TESTING")
	fmt.Fprintln(writer)

	for i, tc := range test.TestCases {
		fmt.Fprintf(writer, "Test case %3d/%3d: %s (size: %d)\n", i+1, len(test.TestCases), tc.Name, len(tc.Data.In))

		success := true

		for j := range tc.Data.In {
			in := tc.Data.In[j]
			expected := tc.Data.Out[j]

			o.In().Push(utils.CleanValue(in))
			actual := o.Out().Pull()

			if !testEqual(expected, actual) {
				fmt.Fprintf(writer, "  expected: %v (%T)\n", expected, expected)
				fmt.Fprintf(writer, "  actual:   %v (%T)\n", actual, actual)

				success = false

				if failFast {
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
	}

	fmt.Fprintln(writer)

	fmt.Fprintln(writer, "SUMMARY")
	fmt.Fprintln(writer)
	fmt.Fprintf(writer, "Tests run: %3d\n", len(test.TestCases))
	fmt.Fprintf(writer, "Succeeded: %3d\n", succs)
	fmt.Fprintf(writer, "Failed:    %3d\n", fails)

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
	if ai, ok := a.(int); ok {
		a = float64(ai)
	}
	if bi, ok := a.(int); ok {
		b = float64(bi)
	}
	return reflect.DeepEqual(a, b)
}