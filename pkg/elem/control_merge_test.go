package elem

import (
	"testing"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/stretchr/testify/require"
)

func Test_CtrlMerge__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocTake := getBuiltinCfg(controlMergeId)
	a.NotNil(ocTake)
}

func Test_CtrlMerge__NoGenerics(t *testing.T) {
	a := assertions.New(t)
	co, err := buildOperator(
		core.InstanceDef{
			Operator: controlMergeId,
		},
	)
	a.Error(err)
	a.Nil(co)
}

func Test_CtrlMerge__InPorts(t *testing.T) {
	a := assertions.New(t)
	to, err := buildOperator(
		core.InstanceDef{
			Operator: controlMergeId,
			Generics: map[string]*core.TypeDef{
				"itemType": {
					Type: "number",
				},
			},
		},
	)
	require.NoError(t, err)
	a.NotNil(to)

	a.Equal(core.TYPE_MAP, to.Main().In().Type())
	a.Equal(core.TYPE_STREAM, to.Main().In().Map("true").Type())
	a.Equal(core.TYPE_STREAM, to.Main().In().Map("false").Type())
	a.Equal(core.TYPE_NUMBER, to.Main().In().Map("true").Stream().Type())
	a.Equal(core.TYPE_NUMBER, to.Main().In().Map("false").Stream().Type())
	a.Equal(core.TYPE_BOOLEAN, to.Delegate("compare").In().Type())
}

func Test_CtrlMerge__OutPorts(t *testing.T) {
	a := assertions.New(t)
	to, err := buildOperator(
		core.InstanceDef{
			Operator: controlMergeId,
			Generics: map[string]*core.TypeDef{
				"itemType": {
					Type: "number",
				},
			},
		},
	)
	require.NoError(t, err)
	a.NotNil(to)

	a.Equal(core.TYPE_STREAM, to.Main().Out().Type())
	a.Equal(core.TYPE_NUMBER, to.Main().Out().Stream().Type())
	a.Equal(core.TYPE_MAP, to.Delegate("compare").Out().Type())
	a.Equal(core.TYPE_NUMBER, to.Delegate("compare").Out().Map("true").Type())
	a.Equal(core.TYPE_NUMBER, to.Delegate("compare").Out().Map("false").Type())
}

func Test_CtrlMerge__Simple1(t *testing.T) {
	a := assertions.New(t)
	to, err := buildOperator(
		core.InstanceDef{
			Operator: controlMergeId,
			Generics: map[string]*core.TypeDef{
				"itemType": {
					Type: "number",
				},
			},
		},
	)
	require.NoError(t, err)
	a.NotNil(to)

	to.Main().Out().Bufferize()
	to.Delegate("compare").Out().Bufferize()
	to.Start()

	// Push data
	to.Main().In().Map("true").Push([]interface{}{1, 2, 3})
	to.Main().In().Map("false").Push([]interface{}{4, 5})

	// Eat BOS

	i := to.Main().Out().Stream().Pull()
	a.True(to.Main().Out().OwnBOS(i))

	// Actual logic

	i = to.Delegate("compare").Out().Pull()
	a.Equal(map[string]interface{}{"true": 1, "false": 4}, i)

	to.Delegate("compare").In().Push(true)

	i = to.Main().Out().Stream().Pull()
	a.Equal(1, i)

	i = to.Delegate("compare").Out().Pull()
	a.Equal(map[string]interface{}{"true": 2, "false": 4}, i)

	to.Delegate("compare").In().Push(false)

	i = to.Main().Out().Stream().Pull()
	a.Equal(4, i)

	i = to.Delegate("compare").Out().Pull()
	a.Equal(map[string]interface{}{"true": 2, "false": 5}, i)

	to.Delegate("compare").In().Push(true)

	i = to.Main().Out().Stream().Pull()
	a.Equal(2, i)

	i = to.Delegate("compare").Out().Pull()
	a.Equal(map[string]interface{}{"true": 3, "false": 5}, i)

	to.Delegate("compare").In().Push(true)

	i = to.Main().Out().Stream().Pull()
	a.Equal(3, i)

	i = to.Main().Out().Stream().Pull()
	a.Equal(5, i)

	i = to.Main().Out().Stream().Pull()
	a.True(to.Main().Out().OwnEOS(i))
}
