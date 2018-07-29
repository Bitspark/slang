package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"testing"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/stretchr/testify/require"
	"github.com/Bitspark/slang/pkg/utils"
)

func TestBuiltin_FileRead__CreatorFuncIsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFileRead := getBuiltinCfg("slang.files.read")
	a.NotNil(ocFileRead)
}

func TestBuiltin_FileRead__InPorts(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.files.read",
		},
	)
	require.NoError(t, err)

	a.NotNil(o.Main().In())
	a.Equal(core.TYPE_STRING, o.Main().In().Type())
}

func TestBuiltin_FileRead__OutPorts(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.files.read",
		},
	)
	require.NoError(t, err)

	a.NotNil(o.Main().Out())
	a.Equal(core.TYPE_MAP, o.Main().Out().Type())
	a.Equal(core.TYPE_BINARY, o.Main().Out().Map("content").Type())
	a.Equal(core.TYPE_STRING, o.Main().Out().Map("error").Type())
}

func TestBuiltin_FileRead__Simple(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.files.read",
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()

	o.Main().In().Push("../../tests/test_data/hello.txt")
	a.Equal(utils.Binary("hello slang"), o.Main().Out().Map("content").Pull())
	a.Nil(o.Main().Out().Map("error").Pull())
}

func TestBuiltin_FileRead__NotFound(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.files.read",
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()

	o.Main().In().Push("./tests/test_data/nonexistentfile")
	a.Nil(o.Main().Out().Map("content").Pull())
	a.NotNil(o.Main().Out().Map("error").Pull())
}
