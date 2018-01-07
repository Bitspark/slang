package builtin

import (
	"slang/core"
	"slang/tests/assertions"
	"testing"
	"github.com/stretchr/testify/require"
)

func TestBuiltin_Fork__CreatorFuncIsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFork := getBuiltinCfg("fork")
	a.NotNil(ocFork)
}

func TestBuiltin_Fork__InPorts(t *testing.T) {
	a := assertions.New(t)

	o, err := MakeOperator(core.InstanceDef{Operator: "fork"})
	require.NoError(t, err)

	a.NotNil(o.In().Stream().Map("i"))
	a.NotNil(o.In().Stream().Map("select"))
}

func TestBuiltin_Fork__OutPorts(t *testing.T) {
	a := assertions.New(t)

	o, err := MakeOperator(core.InstanceDef{Operator: "fork"})
	require.NoError(t, err)

	a.NotNil(o.Out().Map("true").Stream())
	a.NotNil(o.Out().Map("false").Stream())
}

func TestBuiltin_Fork__Correct(t *testing.T) {
	a := assertions.New(t)

	idef := core.PortDef{Type: "primitive"}
	in := core.PortDef{
		Type: "stream",
		Stream: &core.PortDef{
			Type: "map",
			Map:  map[string]core.PortDef{"select": {Type: "boolean"}, "i": idef},
		},
	}
	out := core.PortDef{
		Type: "map",
		Map: map[string]core.PortDef{
			"true":  {Type: "stream", Stream: &idef},
			"false": {Type: "stream", Stream: &idef},
		},
	}

	o, err := MakeOperator(core.InstanceDef{Operator: "fork", In: &in, Out: &out})
	a.NoError(err)

	o.Out().Map("true").Stream().Bufferize()
	o.Out().Map("false").Stream().Bufferize()
	o.Start()

	o.In().Push([]interface{}{
		map[string]interface{}{
			"i":      "hallo",
			"select": true,
		},
		map[string]interface{}{
			"i":      "welt",
			"select": false,
		},
		map[string]interface{}{
			"i":      100,
			"select": true,
		},
		map[string]interface{}{
			"i":      101,
			"select": false,
		},
	})

	a.PortPushes([]interface{}{[]interface{}{"hallo", 100}}, o.Out().Map("true"))
	a.PortPushes([]interface{}{[]interface{}{"welt", 101}}, o.Out().Map("false"))
}

func TestBuiltin_Fork__ComplexItems(t *testing.T) {
	a := assertions.New(t)

	idef := core.PortDef{Type: "map", Map: map[string]core.PortDef{"a": {Type: "primitive"}, "b": {Type: "primitive"}}}
	in := core.PortDef{
		Type: "stream",
		Stream: &core.PortDef{
			Type: "map",
			Map:  map[string]core.PortDef{"select": {Type: "boolean"}, "i": idef},
		},
	}
	out := core.PortDef{
		Type: "map",
		Map: map[string]core.PortDef{
			"true":  {Type: "stream", Stream: &idef},
			"false": {Type: "stream", Stream: &idef},
		},
	}

	o, err := MakeOperator(core.InstanceDef{Operator: "fork", In: &in, Out: &out})
	a.NoError(err)

	o.Out().Map("true").Stream().Bufferize()
	o.Out().Map("false").Stream().Bufferize()
	o.Start()

	o.In().Push([]interface{}{
		map[string]interface{}{
			"i":      map[string]interface{}{"a": "1", "b": "hallo"},
			"select": true,
		},
		map[string]interface{}{
			"i":      map[string]interface{}{"a": "2", "b": "slang"},
			"select": false,
		},
	})

	a.PortPushes([]interface{}{[]interface{}{map[string]interface{}{"a": "1", "b": "hallo"}}}, o.Out().Map("true"))
	a.PortPushes([]interface{}{[]interface{}{map[string]interface{}{"a": "2", "b": "slang"}}}, o.Out().Map("false"))
}
