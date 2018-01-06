package tests

import (
	"io/ioutil"
	"slang"
	"slang/tests/assertions"
	"testing"
)

func TestTestOperator__TrivialTests(t *testing.T) {
	a := assertions.New(t)
	succs, fails, err := slang.TestOperator("test_data/voidOp_test.json", ioutil.Discard, true)
	a.Nil(err)
	a.Equal(1, succs)
	a.Equal(0, fails)
}

func TestTestOperator__SimpleFail(t *testing.T) {
	a := assertions.New(t)
	succs, fails, err := slang.TestOperator("test_data/voidOp_corruptTest.json", ioutil.Discard, true)
	a.Nil(err)
	a.Equal(0, succs)
	a.Equal(1, fails)
}

func TestTestOperator__ComplexTest(t *testing.T) {
	a := assertions.New(t)
	succs, fails, err := slang.TestOperator("test_data/nested_op/usingSubCustomOpDouble_test.json", ioutil.Discard, true)
	a.Nil(err)
	a.Equal(2, succs)
	a.Equal(0, fails)
}

func TestTestOperator__SuiteTests(t *testing.T) {
	a := assertions.New(t)

	succs, fails, err := slang.TestOperator("test_data/suite/polynomial_test.json", ioutil.Discard, false)
	a.Nil(err)
	a.Equal(1, succs)
	a.Equal(0, fails)

	succs, fails, err = slang.TestOperator("test_data/suite/main_test.json", ioutil.Discard, false)
	a.Nil(err)
	a.Equal(2, succs)
	a.Equal(0, fails)
}
