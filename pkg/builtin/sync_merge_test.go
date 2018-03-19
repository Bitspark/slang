package builtin

import (
	"github.com/stretchr/testify/require"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"testing"
)

func TestBuiltin_SyncMerge__CreatorFuncIsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFork := getBuiltinCfg("slang.syncMerge")
	a.NotNil(ocFork)
}

func TestBuiltin_SyncMerge__InPorts(t *testing.T) {
	a := assertions.New(t)

	o, err := MakeOperator(core.InstanceDef{Operator: "slang.syncMerge", Generics: map[string]*core.PortDef{"itemType": {Type: "primitive"}}})
	require.NoError(t, err)

	a.NotNil(o.In().Map("true"))
	a.NotNil(o.In().Map("false"))
	a.NotNil(o.In().Map("select"))
	a.Equal(core.TYPE_PRIMITIVE, o.In().Map("true").Type())
	a.Equal(core.TYPE_PRIMITIVE, o.In().Map("false").Type())
	a.Equal(core.TYPE_BOOLEAN, o.In().Map("select").Type())
}

func TestBuiltin_SyncMerge__OutPorts(t *testing.T) {
	a := assertions.New(t)

	o, err := MakeOperator(core.InstanceDef{Operator: "slang.syncMerge", Generics: map[string]*core.PortDef{"itemType": {Type: "primitive"}}})
	require.NoError(t, err)

	a.NotNil(o.Out())
	a.Equal(core.TYPE_PRIMITIVE, o.Out().Type())
}

func TestBuiltin_SyncMerge__Works(t *testing.T) {
	a := assertions.New(t)

	o, err := MakeOperator(core.InstanceDef{Operator: "slang.syncMerge", Generics: map[string]*core.PortDef{"itemType": {Type: "primitive"}}})
	require.NoError(t, err)

	o.Out().Bufferize()
	o.Start()

	trues := []interface{}{"Roses", 6, false, "Violets", "are", nil, 1, 2, nil, 4}
	falses := []interface{}{nil, "are", "red.", nil, nil, "blue.", "test", nil, 3, nil}
	selects := []interface{}{true, false, false, true, true, false, true, true, false, true}

	for _, v := range trues {
		o.In().Map("true").Push(v)
	}
	for _, v := range falses {
		o.In().Map("false").Push(v)
	}
	for _, v := range selects {
		o.In().Map("select").Push(v)
	}

	a.PortPushesAll([]interface{}{"Roses", "are", "red.", "Violets", "are", "blue.", 1, 2, 3, 4}, o.Out())
}

func TestBuiltin_SyncMerge__ComplexItems(t *testing.T) {
	a := assertions.New(t)
	o, err := MakeOperator(core.InstanceDef{
		Operator: "slang.syncMerge",
		Generics: map[string]*core.PortDef{"itemType": {Type: "map", Map: map[string]*core.PortDef{"red": {Type: "string"}, "blue": {Type: "string"}}}},
	})
	require.NoError(t, err)

	o.Out().Bufferize()
	o.Start()

	trues := []interface{}{
		[]interface{}{},
		map[string]interface{}{
			"red":  "Roses",
			"blue": "Violets",
		},
		map[string]interface{}{
			"red":  "Apples",
			"blue": "Blueberries",
		},
	}
	falses := []interface{}{
		map[string]interface{}{
			"red":  "Red Bull",
			"blue": "Blues",
		},
		nil,
		42,
	}
	selects := []interface{}{false, true, true}

	for _, v := range trues {
		o.In().Map("true").Push(v)
	}
	for _, v := range falses {
		o.In().Map("false").Push(v)
	}
	for _, v := range selects {
		o.In().Map("select").Push(v)
	}

	a.PortPushesAll([]interface{}{
		map[string]interface{}{"red": "Red Bull", "blue": "Blues"},
		map[string]interface{}{"red": "Roses", "blue": "Violets"},
		map[string]interface{}{"red": "Apples", "blue": "Blueberries"},
	}, o.Out())
}
