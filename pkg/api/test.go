package api

import (
	"errors"
	"fmt"
	"io"
	"log"
	"reflect"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/storage"
	"github.com/google/uuid"
)

type TestBench struct {
	stor *storage.Storage
}

func NewTestBench(stor *storage.Storage) *TestBench {
	return &TestBench{stor}
}

// TestOperator reads a file with test data and its corresponding operator and performs the tests.
// It returns the number of failed and succeeded tests and and error in case something went wrong.
// Test failures do not lead to an error. Test failures are printed to the writer.
func (t TestBench) Run(opId uuid.UUID, writer io.Writer, failFast bool) (int, int, error) {
	opDef, err := t.stor.Load(opId)

	if err != nil {
		return 0, 0, err
	}

	if len(opDef.TestCases) == 0 {
		log.Println("no test cases found")
		return 0, 0, nil
	}

	succs := 0
	fails := 0

	for i, tc := range opDef.TestCases {
		if len(tc.Name) < 3 {
			return 0, 0, errors.New("name too short")
		}

		o, err := BuildAndCompile(opId, tc.Generics, tc.Properties, *t.stor)
		if err != nil {
			return 0, 0, err
		}

		fmt.Fprintf(writer, "Test case %3d/%3d: %s (operators: %d, size: %d)\n", i+1, len(opDef.TestCases), tc.Name, len(o.Children()), len(tc.Data.In))

		if err := o.CorrectlyCompiled(); err != nil {
			return 0, 0, err
		}

		o.Main().Out().Bufferize()
		o.Start()

		success := true

		for j := range tc.Data.In {
			in := tc.Data.In[j]
			expected := core.CleanValue(tc.Data.Out[j])

			o.Main().In().Push(core.CleanValue(in))
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
