package tests

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"slang"
	"os"
)

func TestTestOperator__TrivialTests(t *testing.T) {
	a := assert.New(t)
	fails, succs, err := slang.TestOperator("test_data/voidOp_test.json", os.Stdout)
	a.Nil(err)
	a.Equal(succs, 1)
	a.Equal(fails, 0)
}

func TestTestOperator__SimpleFail(t *testing.T) {
	a := assert.New(t)
	fails, succs, err := slang.TestOperator("test_data/voidOp_corruptTest.json", os.Stdout)
	a.Nil(err)
	a.Equal(succs, 0)
	a.Equal(fails, 1)
}