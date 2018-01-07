package builtin

import (
	"slang/core"
	"slang/tests/assertions"
	"testing"
)

func TestBuiltin_Merge_CreatorFuncIsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFork := getCreatorFunc("merge")
	a.NotNil(ocFork)
}

func TestBuiltin_Merge__InPorts(t *testing.T) {
	a := assertions.New(t)

	o, err := createOpMerge(core.InstanceDef{Operator: "merge"})
	a.NoError(err)

	a.NotNil(o.In().Map("true").Stream())
	a.NotNil(o.In().Map("false").Stream())
	a.NotNil(o.In().Map("select").Stream())
}

func TestBuiltin_Merge__OutPorts(t *testing.T) {
	a := assertions.New(t)

	o, err := createOpMerge(core.InstanceDef{Operator: "merge"})
	a.NoError(err)

	a.NotNil(o.Out().Stream())
}

func TestBuiltin_Merge__PortsConfigured_Correctly(t *testing.T) {
	a := assertions.New(t)

	idef := core.PortDef{Type: "map",
		Map: map[string]core.PortDef{
			"a": {Type: "primitive"},
			"b": {Type: "primitive"},
		},
	}
	o, err := createOpMerge(core.InstanceDef{
		Operator: "merge",
		In: &core.PortDef{
			Type: "map",
			Map: map[string]core.PortDef{
				"select": {Type: "stream", Stream: &core.PortDef{Type: "boolean"}},
				"true":   {Type: "stream", Stream: &idef},
				"false":  {Type: "stream", Stream: &idef},
			},
		},
		Out: &core.PortDef{
			Type:   "stream",
			Stream: &idef,
		},
	})
	a.NoError(err)

	a.NotNil(o.In().Map("true").Stream())
	a.NotNil(o.In().Map("false").Stream())
	a.NotNil(o.In().Map("select").Stream())
	a.NotNil(o.In().Map("true").Stream().Map("a"))
	a.NotNil(o.In().Map("true").Stream().Map("b"))

	a.NotNil(o.Out().Stream())
	a.NotNil(o.Out().Stream().Map("a"))
	a.NotNil(o.Out().Stream().Map("b"))
}

func TestBuiltin_Merge__PortsConfigured_Incorrectly_MissingIn(t *testing.T) {
	a := assertions.New(t)

	idef := core.PortDef{Type: "map",
		Map: map[string]core.PortDef{
			"a": {Type: "primitive"},
			"b": {Type: "primitive"},
		},
	}

	_, err := createOpMerge(core.InstanceDef{
		Operator: "merge",
		Out: &core.PortDef{
			Type:   "stream",
			Stream: &idef,
		},
	})
	a.Error(err)
}

func TestBuiltin_Merge__PortsConfigured_Incorrectly_MissingOut(t *testing.T) {
	a := assertions.New(t)

	idef := core.PortDef{Type: "map",
		Map: map[string]core.PortDef{
			"a": {Type: "primitive"},
			"b": {Type: "primitive"},
		},
	}
	_, err := createOpMerge(core.InstanceDef{
		Operator: "merge",
		In: &core.PortDef{
			Type: "map",
			Map: map[string]core.PortDef{
				"select": {Type: "boolean"},
				"true":   {Type: "stream", Stream: &idef},
				"false":  {Type: "stream", Stream: &idef},
			},
		},
	})
	a.Error(err)
}

func TestBuiltin_Merge__Works(t *testing.T) {
	a := assertions.New(t)

	o, err := createOpMerge(core.InstanceDef{Operator: "merge"})
	a.NoError(err)

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

	idef := core.PortDef{Type: "map",
		Map: map[string]core.PortDef{
			"red":  {Type: "primitive"},
			"blue": {Type: "primitive"},
		},
	}

	o, err := createOpMerge(core.InstanceDef{
		Operator: "merge",
		In: &core.PortDef{
			Type: "map",
			Map: map[string]core.PortDef{
				"select": {Type: "stream", Stream: &core.PortDef{Type: "boolean"}},
				"true":   {Type: "stream", Stream: &idef},
				"false":  {Type: "stream", Stream: &idef},
			},
		},
		Out: &core.PortDef{
			Type:   "stream",
			Stream: &idef,
		},
	})
	a.NoError(err)

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
