package tests

import (
	"testing"
	"slang"
	"github.com/stretchr/testify/require"
	"slang/tests/assertions"
	"io/ioutil"
)

func TestStdlibOperators(t *testing.T) {
	a := assertions.New(t)
	succs, fails, err := slang.TestOperator("../slang/math/sum_test.yaml", ioutil.Discard, false)
	require.NoError(t, err)
	a.True(succs > 0)
	a.Equal(0, fails)
}
