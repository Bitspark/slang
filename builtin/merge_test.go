package builtin

import (
	"slang/core"
	"slang/tests/assertions"
	"testing"
	"github.com/stretchr/testify/require"
)

func TestBuiltin_Merge__CreatorFuncIsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFork := getBuiltinCfg("merge")
	a.NotNil(ocFork)
}

func TestBuiltin_Merge__InPorts(t *testing.T) {
	a := assertions.New(t)

	o, err := MakeOperator(&core.InstanceDef{Operator: "merge"})
	require.NoError(t, err)

	a.NotNil(o.In().Map("true").Stream())
	a.NotNil(o.In().Map("false").Stream())
	a.NotNil(o.In().Map("select").Stream())
}

func TestBuiltin_Merge__OutPorts(t *testing.T) {
	a := assertions.New(t)

	o, err := MakeOperator(&core.InstanceDef{Operator: "merge"})
	require.NoError(t, err)

	a.NotNil(o.Out().Stream())
}

func TestBuiltin_Merge__Works(t *testing.T) {
	a := assertions.New(t)

	o, err := MakeOperator(&core.InstanceDef{Operator: "merge"})
	require.NoError(t, err)

	o.Out().Stream().Bufferize()
	o.Start()

	trues := []interface{}{"Roses", "Violets", "are", 1, 2, 4}
	falses := []interface{}{"are", "red.", "blue.", 3}
	selects := []interface{}{true, false, false, true, true, false, true, true, false, true}

	o.In().Map("true").Push(trues)
	o.In().Map("false").Push(falses)
	o.In().Map("select").Push(selects)

	a.PortPushes([]interface{}{[]interface{}{"Roses", "are", "red.", "Violets", "are", "blue.", 1, 2, 3, 4}}, o.Out())
}

func TestBuiltin_Merge__ComplexItems(t *testing.T) {
	a := assertions.New(t)
	o, err := MakeOperator(&core.InstanceDef{
		Operator: "merge",
	})
	require.NoError(t, err)

	o.Out().Stream().Bufferize()
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

	o.In().Map("true").Push(trues)
	o.In().Map("false").Push(falses)
	o.In().Map("select").Push(selects)

	a.PortPushes([]interface{}{[]interface{}{
		map[string]interface{}{"red": "Red Bull", "blue": "Blues"},
		map[string]interface{}{"red": "Roses", "blue": "Violets"},
		map[string]interface{}{"red": "Apples", "blue": "Blueberries"},
	}}, o.Out())
}
