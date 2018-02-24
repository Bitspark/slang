package builtin

import (
	"testing"
	"slang/tests/assertions"
	"slang/core"
	"github.com/stretchr/testify/require"
)

func TestOperator_Take__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocTake := getBuiltinCfg("slang.take")
	a.NotNil(ocTake)
}

func TestOperator_Take__NoGenerics(t *testing.T) {
	a := assertions.New(t)
	co, err := MakeOperator(
		core.InstanceDef{
			Operator: "slang.take",
		},
	)
	a.Error(err)
	a.Nil(co)
}

func TestOperator_Take__InPorts(t *testing.T) {
	a := assertions.New(t)
	to, err := MakeOperator(
		core.InstanceDef{
			Operator: "slang.take",
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
	a.Equal(core.TYPE_STREAM, to.Delegate("compare").In().Type())
	a.Equal(core.TYPE_NUMBER, to.In().Map("true").Stream().Type())
	a.Equal(core.TYPE_NUMBER, to.In().Map("false").Stream().Type())
	a.Equal(core.TYPE_BOOLEAN, to.Delegate("compare").In().Stream().Type())
}

func TestOperator_Take__OutPorts(t *testing.T) {
	a := assertions.New(t)
	to, err := MakeOperator(
		core.InstanceDef{
			Operator: "slang.take",
			Generics: map[string]*core.PortDef{
				"itemType": {
					Type: "number",
				},
			},
		},
	)
	require.NoError(t, err)
	a.NotNil(to)

	a.Equal(core.TYPE_STREAM, to.Out().Type())
	a.Equal(core.TYPE_NUMBER, to.Out().Stream().Type())
	a.Equal(core.TYPE_STREAM, to.Delegate("compare").Out().Type())
	a.Equal(core.TYPE_MAP, to.Delegate("compare").Out().Stream().Type())
	a.Equal(core.TYPE_NUMBER, to.Delegate("compare").Out().Stream().Map("true").Type())
	a.Equal(core.TYPE_NUMBER, to.Delegate("compare").Out().Stream().Map("false").Type())
}

func TestOperator_Take__Simple1(t *testing.T) {
	a := assertions.New(t)
	to, err := MakeOperator(
		core.InstanceDef{
			Operator: "slang.take",
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
	to.Delegate("compare").Out().Bufferize()
	to.Start()

	// Push data
	to.In().Map("true").Push([]interface{}{1, 2, 3})
	to.In().Map("false").Push([]interface{}{4, 5})

	// Push BOS to select stream
	to.Delegate("compare").In().PushBOS()

	// Eat BOS

	i := to.Out().Stream().Pull()
	a.True(to.Out().OwnBOS(i))

	i = to.Delegate("compare").Out().Stream().Pull()
	a.True(to.Delegate("compare").Out().OwnBOS(i))

	// Actual logic

	i = to.Delegate("compare").Out().Stream().Pull()
	a.Equal(map[string]interface{}{"true": 1, "false": 4}, i)

	to.Delegate("compare").In().Stream().Push(true)

	i = to.Out().Stream().Pull()
	a.Equal(1, i)


	i = to.Delegate("compare").Out().Pull()
	a.Equal(map[string]interface{}{"true": 2, "false": 4}, i)

	to.Delegate("compare").In().Stream().Push(false)

	i = to.Out().Stream().Pull()
	a.Equal(4, i)


	i = to.Delegate("compare").Out().Pull()
	a.Equal(map[string]interface{}{"true": 2, "false": 5}, i)

	to.Delegate("compare").In().Stream().Push(true)

	i = to.Out().Stream().Pull()
	a.Equal(2, i)


	i = to.Delegate("compare").Out().Pull()
	a.Equal(map[string]interface{}{"true": 3, "false": 5}, i)

	to.Delegate("compare").In().Stream().Push(true)

	i = to.Out().Stream().Pull()
	a.Equal(3, i)

	i = to.Out().Stream().Pull()
	a.Equal(5, i)

	// Eat EOS

	i = to.Delegate("compare").Out().Stream().Pull()
	a.True(to.Delegate("compare").Out().OwnEOS(i))

	// Push EOS

	to.Delegate("compare").In().PushEOS()

	i = to.Out().Stream().Pull()
	a.True(to.Out().OwnEOS(i))
}
