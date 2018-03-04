package builtin

import (
	"github.com/stretchr/testify/require"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"testing"
)

func TestOperatorCreator__Aggregate__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocAgg := getBuiltinCfg("slang.aggregate")
	a.NotNil(ocAgg)
}

func TestBuiltinAggregate__PassOtherMarkers(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	ao, err := MakeOperator(
		core.InstanceDef{
			Operator: "slang.aggregate",
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
		nil,
		core.PortDef{Type: "stream",
			Stream: &core.PortDef{Type: "map",
				Map: map[string]*core.PortDef{
					"init":  {Type: "number"},
					"items": {Type: "stream", Stream: &core.PortDef{Type: "number"}}}}},
		core.PortDef{Type: "stream",
			Stream: &core.PortDef{Type: "number"}},
		nil)
	require.NoError(t, err)
	a.NotNil(do)

	ao.SetParent(do)

	r.NoError(do.In().Stream().Map("init").Connect(ao.In().Map("init")))
	r.NoError(do.In().Stream().Map("items").Connect(ao.In().Map("items")))
	r.NoError(ao.Delegate("iteration").Out().Stream().Map("state").Connect(ao.Delegate("iteration").In().Stream()))
	r.NoError(ao.Out().Connect(do.Out().Stream()))

	do.Out().Bufferize()

	do.Start()

	do.In().Push([]interface{}{map[string]interface{}{"init": 0.0, "items": []interface{}{}}})
	a.PortPushes([]interface{}{[]interface{}{0.0}}, do.Out())
}

func TestBuiltinAggregate__SimpleLoop(t *testing.T) {
	a := assertions.New(t)
	ao, err := MakeOperator(
		core.InstanceDef{
			Operator: "slang.aggregate",
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
	fo, err := core.NewOperator("add", func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
		for true {
			i := in.Pull()
			m, ok := i.(map[string]interface{})
			if !ok {
				out.Push(i)
			} else {
				out.Push(m["state"].(float64) + m["item"].(float64))
			}
		}
	},
		nil,
		core.PortDef{Type: "map", Map: map[string]*core.PortDef{"state": {Type: "number"}, "item": {Type: "number"}}},
		core.PortDef{Type: "number"},
		nil)
	require.NoError(t, err)

	// Connect
	require.NoError(t, ao.Delegate("iteration").Out().Stream().Connect(fo.In()))
	require.NoError(t, fo.Out().Connect(ao.Delegate("iteration").In().Stream()))

	ao.Out().Bufferize()

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

	a.PortPushes([]interface{}{6.0, 20.0, 999.0, 10.0}, ao.Out())
}

func TestBuiltinAggregate__PassMarkers(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := core.NewOperator(
		"test",
		func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {},
		nil,
		core.PortDef{Type: "trigger"},
		core.PortDef{Type: "map",
			Map: map[string]*core.PortDef{
				"init":  {Type: "number"},
				"items": {Type: "stream", Stream: &core.PortDef{Type: "number"}},
			},
		},
		nil,
	)
	r.NoError(err)
	a.NotNil(o)

	ao, err := MakeOperator(
		core.InstanceDef{
			Name:     "testOp",
			Operator: "slang.aggregate",
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
	r.NoError(err)
	a.NotNil(ao)

	err = o.Out().Connect(ao.In())
	r.NoError(err)

	ao.Out().Bufferize()
	ao.Delegate("iteration").Out().Bufferize()

	ao.Start()

	a.Equal(o.Out().Map("items"), ao.Delegate("iteration").Out().StreamSource())

	o.Out().Push(map[string]interface{}{"init": 0, "items": []interface{}{1, 2, 3}})

	i := ao.Delegate("iteration").Out().Stream().Map("item").Pull()
	a.True(ao.Delegate("iteration").Out().OwnBOS(i))
	a.True(o.Out().Map("items").OwnBOS(i))
}
