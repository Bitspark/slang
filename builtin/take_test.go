package builtin

import (
	"testing"
	"slang/tests/assertions"
	"slang/core"
	"github.com/stretchr/testify/require"
)

func TestOperator_Take__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocTake := getBuiltinCfg("take")
	a.NotNil(ocTake)
}

func TestOperator_Take__NoGenerics(t *testing.T) {
	a := assertions.New(t)
	co, err := MakeOperator(
		core.InstanceDef{
			Operator: "take",
		},
	)
	a.Error(err)
	a.Nil(co)
}

func TestOperator_Take__InPorts(t *testing.T) {
	a := assertions.New(t)
	to, err := MakeOperator(
		core.InstanceDef{
			Operator: "take",
			Generics: map[string]*core.PortDef{
				"itemType": {
					Type: "number",
				},
			},
		},
	)
	require.NoError(t, err)
	a.NotNil(to)

	a.Equal(core.TYPE_MAP, to.In().Type())
	a.Equal(core.TYPE_STREAM, to.In().Map("true").Type())
	a.Equal(core.TYPE_STREAM, to.In().Map("false").Type())
	a.Equal(core.TYPE_STREAM, to.In().Map("select").Type())
	a.Equal(core.TYPE_NUMBER, to.In().Map("true").Stream().Type())
	a.Equal(core.TYPE_NUMBER, to.In().Map("false").Stream().Type())
	a.Equal(core.TYPE_BOOLEAN, to.In().Map("select").Stream().Type())
}

func TestOperator_Take__OutPorts(t *testing.T) {
	a := assertions.New(t)
	to, err := MakeOperator(
		core.InstanceDef{
			Operator: "take",
			Generics: map[string]*core.PortDef{
				"itemType": {
					Type: "number",
				},
			},
		},
	)
	require.NoError(t, err)
	a.NotNil(to)

	a.Equal(core.TYPE_MAP, to.Out().Type())
	a.Equal(core.TYPE_STREAM, to.Out().Map("result").Type())
	a.Equal(core.TYPE_NUMBER, to.Out().Map("result").Stream().Type())
	a.Equal(core.TYPE_STREAM, to.Out().Map("compare").Type())
	a.Equal(core.TYPE_MAP, to.Out().Map("compare").Stream().Type())
	a.Equal(core.TYPE_NUMBER, to.Out().Map("compare").Stream().Map("true").Type())
	a.Equal(core.TYPE_NUMBER, to.Out().Map("compare").Stream().Map("false").Type())
}

func TestOperator_Take__Simple1(t *testing.T) {
	a := assertions.New(t)
	to, err := MakeOperator(
		core.InstanceDef{
			Operator: "take",
			Generics: map[string]*core.PortDef{
				"itemType": {
					Type: "number",
				},
			},
		},
	)
	require.NoError(t, err)
	a.NotNil(to)

	to.Out().Bufferize()
	to.Start()

	// Push data
	to.In().Map("true").Push([]interface{}{1, 2, 3})
	to.In().Map("false").Push([]interface{}{4, 5})

	// Push BOS to select stream
	to.In().Map("select").PushBOS()

	// Eat BOS

	i := to.Out().Map("result").Stream().Pull()
	a.True(to.Out().Map("result").OwnBOS(i))

	i = to.Out().Map("compare").Stream().Pull()
	a.True(to.Out().Map("compare").OwnBOS(i))

	// Actual logic

	i = to.Out().Map("compare").Stream().Pull()
	a.Equal(map[string]interface{}{"true": 1, "false": 4}, i)

	to.In().Map("select").Stream().Push(true)

	i = to.Out().Map("result").Stream().Pull()
	a.Equal(1, i)


	i = to.Out().Map("compare").Pull()
	a.Equal(map[string]interface{}{"true": 2, "false": 4}, i)

	to.In().Map("select").Stream().Push(false)

	i = to.Out().Map("result").Stream().Pull()
	a.Equal(4, i)


	i = to.Out().Map("compare").Pull()
	a.Equal(map[string]interface{}{"true": 2, "false": 5}, i)

	to.In().Map("select").Stream().Push(true)

	i = to.Out().Map("result").Stream().Pull()
	a.Equal(2, i)


	i = to.Out().Map("compare").Pull()
	a.Equal(map[string]interface{}{"true": 3, "false": 5}, i)

	to.In().Map("select").Stream().Push(true)

	i = to.Out().Map("result").Stream().Pull()
	a.Equal(3, i)

	i = to.Out().Map("result").Stream().Pull()
	a.Equal(5, i)

	// Eat EOS

	i = to.Out().Map("compare").Stream().Pull()
	a.True(to.Out().Map("compare").OwnEOS(i))

	// Push EOS

	to.In().Map("select").PushEOS()

	i = to.Out().Map("result").Stream().Pull()
	a.True(to.Out().Map("result").OwnEOS(i))
}
