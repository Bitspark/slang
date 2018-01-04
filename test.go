package slang

import (
	"io"
	"io/ioutil"
	"encoding/json"
	"errors"
)

type TestCaseDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Data struct {
		In  []interface{} `json:"in"`
		Out []interface{} `json:"out"`
	}
}

type TestDef struct {
	OperatorFile string        `json:"operatorFile"`
	Description  string        `json:"description"`
	TestCases    []TestCaseDef `json:"testCases"`
	valid        bool
}

// TestOperator reads a file with test data and its corresponding operator and performs the tests.
// It returns the number of failed and succeeded tests and and error in case something went wrong.
// Test failures do not lead to an error. Test failures are printed to the writer.
func TestOperator(testDataPath string, writer io.Writer) (int, int, error) {
	b, err := ioutil.ReadFile(testDataPath)

	if err != nil {
		return 0, 0, err
	}

	test := TestDef{}
	json.Unmarshal(b, &test)

	if !test.Valid() {
		err := test.Validate()
		if err != nil {
			return 0, 0, err
		}
	}

	return 0, 0, nil
}

func (t TestDef) Validate() error {
	if len(t.OperatorFile) == 0 {
		return errors.New("no operator file given")
	}

	t.valid = true
	return nil
}

func (t TestDef) Valid() bool {
	return t.valid
}
