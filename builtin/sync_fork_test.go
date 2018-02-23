package builtin

import (
	"github.com/stretchr/testify/require"
	"slang/core"
	"slang/tests/assertions"
	"testing"
)

func TestBuiltin_SyncFork__CreatorFuncIsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFork := getBuiltinCfg("syncFork")
	a.NotNil(ocFork)
}

func TestBuiltin_SyncFork__InPorts(t *testing.T) {
	a := assertions.New(t)

	o, err := MakeOperator(
		core.InstanceDef{
			Operator: "syncFork",
			Generics: map[string]*core.PortDef{
				"itemType": {
					Type: "primitive",
				},
			},
		},
	)
	require.NoError(t, err)

	a.NotNil(o.In().Map("item"))
	a.NotNil(o.In().Map("select"))
	a.Equal(core.TYPE_PRIMITIVE, o.In().Map("item").Type())
	a.Equal(core.TYPE_BOOLEAN, o.In().Map("select").Type())
}

func TestBuiltin_SyncFork__OutPorts(t *testing.T) {
	a := assertions.New(t)

	o, err := MakeOperator(
		core.InstanceDef{
			Operator: "syncFork",
			Generics: map[string]*core.PortDef{
				"itemType": {
					Type: "primitive",
				},
			},
		},
	)
	require.NoError(t, err)

	a.NotNil(o.Out().Map("true"))
	a.NotNil(o.Out().Map("false"))
	a.Equal(core.TYPE_PRIMITIVE, o.Out().Map("true").Type())
	a.Equal(core.TYPE_PRIMITIVE, o.Out().Map("false").Type())
}

func TestBuiltin_SyncFork__Correct(t *testing.T) {
	a := assertions.New(t)

	o, err := MakeOperator(
		core.InstanceDef{
			Operator: "syncFork",
			Generics: map[string]*core.PortDef{
				"itemType": {
					Type: "primitive",
				},
			},
		},
	)
	require.NoError(t, err)

	o.Out().Bufferize()
	o.Start()

	o.In().Push(
		map[string]interface{}{
			"item":   "hallo",
			"select": true,
		})
	o.In().Push(
		map[string]interface{}{
			"item":   "welt",
			"select": false,
		})
	o.In().Push(
		map[string]interface{}{
			"item":   100,
			"select": true,
		})
	o.In().Push(
		map[string]interface{}{
			"item":   101,
			"select": false,
		})

	a.PortPushes([]interface{}{"hallo", nil, 100, nil}, o.Out().Map("true"))
	a.PortPushes([]interface{}{nil, "welt", nil, 101}, o.Out().Map("false"))
}

func TestBuiltin_SyncFork__ComplexItems(t *testing.T) {
	a := assertions.New(t)
	o, err := MakeOperator(
		core.InstanceDef{
			Operator: "syncFork",
			Generics: map[string]*core.PortDef{
				"itemType": {
					Type: "map",
					Map: map[string]*core.PortDef{
						"a": {Type: "number"},
						"b": {Type: "string"},
					},
				},
			},
		},
	)
	a.NoError(err)

	o.Out().Bufferize()
	o.Start()

	o.In().Push(
		map[string]interface{}{
			"item":   map[string]interface{}{"a": "1", "b": "hallo"},
			"select": true,
		})
	o.In().Push(
		map[string]interface{}{
			"item":   map[string]interface{}{"a": "2", "b": "slang"},
			"select": false,
		})

	a.PortPushes([]interface{}{map[string]interface{}{"a": "1", "b": "hallo"}, map[string]interface{}{"a": nil, "b": nil}}, o.Out().Map("true"))
	a.PortPushes([]interface{}{map[string]interface{}{"a": nil, "b": nil}, map[string]interface{}{"a": "2", "b": "slang"}}, o.Out().Map("false"))
}
