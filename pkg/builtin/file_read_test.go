package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
	"testing"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/stretchr/testify/require"
)

func TestBuiltin_FileRead__CreatorFuncIsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFileRead := getBuiltinCfg("slang.files.read")
	a.NotNil(ocFileRead)
}

func TestBuiltin_FileRead__InPorts(t *testing.T) {
	a := assertions.New(t)

	o, err := MakeOperator(
		core.InstanceDef{
			Operator: "slang.files.read",
		},
	)
	require.NoError(t, err)

	a.NotNil(o.In())
	a.Equal(core.TYPE_STRING, o.In().Type())
}

func TestBuiltin_FileRead__OutPorts(t *testing.T) {
	a := assertions.New(t)

	o, err := MakeOperator(
		core.InstanceDef{
			Operator: "slang.files.read",
		},
	)
	require.NoError(t, err)

	a.NotNil(o.Out())
	a.Equal(core.TYPE_MAP, o.Out().Type())
	a.Equal(core.TYPE_BINARY, o.Out().Map("content").Type())
	a.Equal(core.TYPE_STRING, o.Out().Map("error").Type())
}

func TestBuiltin_FileRead__Simple(t *testing.T) {
	a := assertions.New(t)

	o, err := MakeOperator(
		core.InstanceDef{
			Operator: "slang.files.read",
		},
	)
	require.NoError(t, err)

	o.Out().Bufferize()
	o.Start()

	o.In().Push("../../tests/test_data/hello.txt")
	a.Equal([]byte("hello slang"), o.Out().Map("content").Pull())
	a.Nil(o.Out().Map("error").Pull())
}

func TestBuiltin_FileRead__NotFound(t *testing.T) {
	a := assertions.New(t)

	o, err := MakeOperator(
		core.InstanceDef{
			Operator: "slang.files.read",
		},
	)
	require.NoError(t, err)

	o.Out().Bufferize()
	o.Start()

	o.In().Push("./tests/test_data/nonexistentfile")
	a.Nil(o.Out().Map("content").Pull())
	a.NotNil(o.Out().Map("error").Pull())
}
