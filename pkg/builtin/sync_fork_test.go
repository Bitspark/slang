package builtin

import (
	"github.com/stretchr/testify/require"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"testing"
)

func TestBuiltin_SyncFork__CreatorFuncIsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFork := getBuiltinCfg("slang.syncFork")
	a.NotNil(ocFork)
}

func TestBuiltin_SyncFork__InPorts(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.syncFork",
			Generics: map[string]*core.TypeDef{
				"itemType": {
					Type: "primitive",
				},
			},
		},
	)
	require.NoError(t, err)

	a.NotNil(o.Main().In().Map("item"))
	a.NotNil(o.Main().In().Map("select"))
	a.Equal(core.TYPE_PRIMITIVE, o.Main().In().Map("item").Type())
	a.Equal(core.TYPE_BOOLEAN, o.Main().In().Map("select").Type())
}

func TestBuiltin_SyncFork__OutPorts(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.syncFork",
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

func TestBuiltin_SyncFork__Correct(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.syncFork",
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
			"select": true,
		})
	o.Main().In().Push(
		map[string]interface{}{
			"item":   "welt",
			"select": false,
		})
	o.Main().In().Push(
		map[string]interface{}{
			"item":   100,
			"select": true,
		})
	o.Main().In().Push(
		map[string]interface{}{
			"item":   101,
			"select": false,
		})

	a.PortPushesAll([]interface{}{"hallo", nil, 100, nil}, o.Main().Out().Map("true"))
	a.PortPushesAll([]interface{}{nil, "welt", nil, 101}, o.Main().Out().Map("false"))
}

func TestBuiltin_SyncFork__ComplexItems(t *testing.T) {
	a := assertions.New(t)
	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.syncFork",
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
			"select": true,
		})
	o.Main().In().Push(
		map[string]interface{}{
			"item":   map[string]interface{}{"a": "2", "b": "slang"},
			"select": false,
		})

	a.PortPushesAll([]interface{}{map[string]interface{}{"a": "1", "b": "hallo"}, map[string]interface{}{"a": nil, "b": nil}}, o.Main().Out().Map("true"))
	a.PortPushesAll([]interface{}{map[string]interface{}{"a": nil, "b": nil}, map[string]interface{}{"a": "2", "b": "slang"}}, o.Main().Out().Map("false"))
}
