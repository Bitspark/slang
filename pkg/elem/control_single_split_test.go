package elem

import (
	"github.com/stretchr/testify/require"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"testing"
)

func Test_CtrlSingleSplit__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFork := getBuiltinCfg("slang.control.SingleSplit")
	a.NotNil(ocFork)
}

func Test_CtrlSingleSplit__InPorts(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.control.SingleSplit",
			Generics: map[string]*core.TypeDef{
				"itemType": {
					Type: "primitive",
				},
			},
		},
	)
	require.NoError(t, err)

	a.NotNil(o.Main().In().Map("item"))
	a.NotNil(o.Main().In().Map("control"))
	a.Equal(core.TYPE_PRIMITIVE, o.Main().In().Map("item").Type())
	a.Equal(core.TYPE_BOOLEAN, o.Main().In().Map("control").Type())
}

func Test_CtrlSingleSplit__OutPorts(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.control.SingleSplit",
			Generics: map[string]*core.TypeDef{
				"itemType": {
					Type: "primitive",
				},
			},
		},
	)
	require.NoError(t, err)

	a.NotNil(o.Main().Out().Map("true"))
	a.NotNil(o.Main().Out().Map("false"))
	a.Equal(core.TYPE_PRIMITIVE, o.Main().Out().Map("true").Type())
	a.Equal(core.TYPE_PRIMITIVE, o.Main().Out().Map("false").Type())
}

func Test_CtrlSingleSplit__Correct(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.control.SingleSplit",
			Generics: map[string]*core.TypeDef{
				"itemType": {
					Type: "primitive",
				},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()

	o.Main().In().Push(
		map[string]interface{}{
			"item":   "hallo",
			"control": true,
		})
	o.Main().In().Push(
		map[string]interface{}{
			"item":   "welt",
			"control": false,
		})
	o.Main().In().Push(
		map[string]interface{}{
			"item":   100,
			"control": true,
		})
	o.Main().In().Push(
		map[string]interface{}{
			"item":   101,
			"control": false,
		})

	a.PortPushesAll([]interface{}{"hallo", nil, 100, nil}, o.Main().Out().Map("true"))
	a.PortPushesAll([]interface{}{nil, "welt", nil, 101}, o.Main().Out().Map("false"))
}

func Test_CtrlSingleSplit__ComplexItems(t *testing.T) {
	a := assertions.New(t)
	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.control.SingleSplit",
			Generics: map[string]*core.TypeDef{
				"itemType": {
					Type: "map",
					Map: map[string]*core.TypeDef{
						"a": {Type: "number"},
						"b": {Type: "string"},
					},
				},
			},
		},
	)
	a.NoError(err)

	o.Main().Out().Bufferize()
	o.Start()

	o.Main().In().Push(
		map[string]interface{}{
			"item":   map[string]interface{}{"a": "1", "b": "hallo"},
			"control": true,
		})
	o.Main().In().Push(
		map[string]interface{}{
			"item":   map[string]interface{}{"a": "2", "b": "slang"},
			"control": false,
		})

	a.PortPushesAll([]interface{}{map[string]interface{}{"a": "1", "b": "hallo"}, map[string]interface{}{"a": nil, "b": nil}}, o.Main().Out().Map("true"))
	a.PortPushesAll([]interface{}{map[string]interface{}{"a": nil, "b": nil}, map[string]interface{}{"a": "2", "b": "slang"}}, o.Main().Out().Map("false"))
}
