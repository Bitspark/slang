package builtin

import (
	"testing"
	"slang/tests/assertions"
	"slang/core"
	"github.com/stretchr/testify/require"
)

func TestOperatorCreator__Agg__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocAgg := getBuiltinCfg("agg")
	a.NotNil(ocAgg)
}

func TestBuiltinAgg__PassOtherMarkers(t *testing.T) {
	a := assertions.New(t)
	ao, err := MakeOperator(
		core.InstanceDef{
			Operator: "agg",
			Generics: map[string]*core.PortDef{
				"itemType": {
					Type: "number",
				},
				"stateType": {
					Type: "number",
				},
			},
		},
	)
	require.NoError(t, err)
	a.NotNil(ao)

	do, err := core.NewOperator(
		"wrapper",
		nil,
		core.PortDef{Type: "stream",
			Stream: &core.PortDef{Type: "map",
				Map: map[string]*core.PortDef{
					"init":  {Type: "number"},
					"items": {Type: "stream", Stream: &core.PortDef{Type: "number"}}}}},
		core.PortDef{Type: "stream",
			Stream: &core.PortDef{Type: "number"}})
	require.NoError(t, err)
	a.NotNil(do)

	ao.SetParent(do)

	ao.Out().Map("iteration").Stream().Map("state").Connect(ao.In().Map("state").Stream())

	require.NoError(t, do.In().Stream().Map("init").Connect(ao.In().Map("init")))
	require.NoError(t, do.In().Stream().Map("items").Connect(ao.In().Map("items")))
	require.NoError(t, ao.Out().Map("end").Connect(do.Out().Stream()))

	do.Out().Stream().Bufferize()

	do.Start()

	do.In().Push([]interface{}{map[string]interface{}{"init": 0.0, "items": []interface{}{}}})
	a.PortPushes([]interface{}{[]interface{}{0.0}}, do.Out())
}

func TestBuiltinAgg__SimpleLoop(t *testing.T) {
	a := assertions.New(t)
	ao, err := MakeOperator(
		core.InstanceDef{
			Operator: "agg",
			Generics: map[string]*core.PortDef{
				"itemType": {
					Type: "number",
				},
				"stateType": {
					Type: "number",
				},
			},
		},
	)
	require.NoError(t, err)
	a.NotNil(ao)

	// Add function operator
	fo, err := core.NewOperator("add", func(in, out *core.Port, store interface{}) {
		for true {
			i := in.Pull()
			m, ok := i.(map[string]interface{})
			if !ok {
				out.Push(i)
			} else {
				out.Push(m["state"].(float64) + m["i"].(float64))
			}
		}
	},
		core.PortDef{Type: "map", Map: map[string]*core.PortDef{"state": {Type: "number"}, "i": {Type: "number"}}},
		core.PortDef{Type: "number"})
	require.NoError(t, err)

	// Connect
	require.NoError(t, ao.Out().Map("iteration").Stream().Connect(fo.In()))
	require.NoError(t, fo.Out().Connect(ao.In().Map("state").Stream()))

	ao.Out().Map("end").Bufferize()

	ao.In().Map("init").Push(0.0)
	ao.In().Map("init").Push(8.0)
	ao.In().Map("init").Push(999.0)
	ao.In().Map("init").Push(4.0)
	ao.In().Map("items").Push([]interface{}{1.0, 2.0, 3.0})
	ao.In().Map("items").Push([]interface{}{2.0, 4.0, 6.0})
	ao.In().Map("items").Push([]interface{}{})
	ao.In().Map("items").Push([]interface{}{1.0, 2.0, 3.0})

	ao.Start()
	fo.Start()

	a.PortPushes([]interface{}{6.0, 20.0, 999.0, 10.0}, ao.Out().Map("end"))
}
