package elem

import (
	"github.com/stretchr/testify/require"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"testing"
)

func TestBuiltin_Merge__CreatorFuncIsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFork := getBuiltinCfg("slang.merge")
	a.NotNil(ocFork)
}

func TestBuiltin_Merge__InPorts(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(core.InstanceDef{Operator: "slang.merge", Generics: map[string]*core.TypeDef{"itemType": {Type: "primitive"}}})
	require.NoError(t, err)

	a.NotNil(o.Main().In().Map("true").Stream())
	a.NotNil(o.Main().In().Map("false").Stream())
	a.NotNil(o.Main().In().Map("select").Stream())
}

func TestBuiltin_Merge__OutPorts(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(core.InstanceDef{Operator: "slang.merge", Generics: map[string]*core.TypeDef{"itemType": {Type: "primitive"}}})
	require.NoError(t, err)

	a.NotNil(o.Main().Out().Stream())
}

func TestBuiltin_Merge__Works(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(core.InstanceDef{Operator: "slang.merge", Generics: map[string]*core.TypeDef{"itemType": {Type: "primitive"}}})
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()

	trues := []interface{}{"Roses", "Violets", "are", 1, 2, 4}
	falses := []interface{}{"are", "red.", "blue.", 3}
	selects := []interface{}{true, false, false, true, true, false, true, true, false, true}

	o.Main().In().Map("true").Push(trues)
	o.Main().In().Map("false").Push(falses)
	o.Main().In().Map("select").Push(selects)

	a.PortPushesAll([]interface{}{[]interface{}{"Roses", "are", "red.", "Violets", "are", "blue.", 1, 2, 3, 4}}, o.Main().Out())
}

func TestBuiltin_Merge__ComplexItems(t *testing.T) {
	a := assertions.New(t)
	o, err := buildOperator(core.InstanceDef{
		Operator: "slang.merge",
		Generics: map[string]*core.TypeDef{"itemType": {Type: "map", Map: map[string]*core.TypeDef{"red": {Type: "string"}, "blue": {Type: "string"}}}},
	})
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()

	trues := []interface{}{
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
	}
	selects := []interface{}{false, true, true}

	o.Main().In().Map("true").Push(trues)
	o.Main().In().Map("false").Push(falses)
	o.Main().In().Map("select").Push(selects)

	a.PortPushesAll([]interface{}{[]interface{}{
		map[string]interface{}{"red": "Red Bull", "blue": "Blues"},
		map[string]interface{}{"red": "Roses", "blue": "Violets"},
		map[string]interface{}{"red": "Apples", "blue": "Blueberries"},
	}}, o.Main().Out())
}
