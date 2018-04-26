package builtin

import (
	"github.com/Bitspark/slang/tests/assertions"
	"testing"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/stretchr/testify/require"
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
			Generics: map[string]*core.TypeDef{
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
		nil,
		nil,
		core.OperatorDef{
			ServiceDefs: map[string]*core.ServiceDef{
				"main": {
					In: core.TypeDef{Type: "stream",
						Stream: &core.TypeDef{Type: "map",
							Map: map[string]*core.TypeDef{
								"init":  {Type: "number"},
								"items": {Type: "stream", Stream: &core.TypeDef{Type: "number"}}}}},
					Out: core.TypeDef{Type: "stream",
						Stream: &core.TypeDef{Type: "number"}},
				},
			},
		},
	)
	require.NoError(t, err)
	a.NotNil(do)

	ao.SetParent(do)

	r.NoError(do.Main().In().Stream().Map("init").Connect(ao.Main().In().Map("init")))
	r.NoError(do.Main().In().Stream().Map("items").Connect(ao.Main().In().Map("items")))
	r.NoError(ao.Delegate("iteration").Out().Stream().Map("state").Connect(ao.Delegate("iteration").In().Stream()))
	r.NoError(ao.Main().Out().Connect(do.Main().Out().Stream()))

	do.Main().Out().Bufferize()

	do.Start()

	do.Main().In().Push([]interface{}{map[string]interface{}{"init": 0.0, "items": []interface{}{}}})
	a.PortPushesAll([]interface{}{[]interface{}{0.0}}, do.Main().Out())
}

func TestBuiltinAggregate__SimpleLoop(t *testing.T) {
	a := assertions.New(t)
	ao, err := MakeOperator(
		core.InstanceDef{
			Operator: "slang.aggregate",
			Generics: map[string]*core.TypeDef{
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
	fo, err := core.NewOperator(
		"add",
		func(op *core.Operator) {
			in := op.Main().In()
			out := op.Main().Out()
			for {
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
		nil,
		nil,
		core.OperatorDef{
			ServiceDefs: map[string]*core.ServiceDef{
				"main": {
					In:  core.TypeDef{Type: "map", Map: map[string]*core.TypeDef{"state": {Type: "number"}, "item": {Type: "number"}}},
					Out: core.TypeDef{Type: "number"},
				},
			},
		},
	)
	require.NoError(t, err)

	// Connect
	require.NoError(t, ao.Delegate("iteration").Out().Stream().Connect(fo.Main().In()))
	require.NoError(t, fo.Main().Out().Connect(ao.Delegate("iteration").In().Stream()))

	ao.Main().Out().Bufferize()

	ao.Main().In().Map("init").Push(0.0)
	ao.Main().In().Map("init").Push(8.0)
	ao.Main().In().Map("init").Push(999.0)
	ao.Main().In().Map("init").Push(4.0)
	ao.Main().In().Map("items").Push([]interface{}{1.0, 2.0, 3.0})
	ao.Main().In().Map("items").Push([]interface{}{2.0, 4.0, 6.0})
	ao.Main().In().Map("items").Push([]interface{}{})
	ao.Main().In().Map("items").Push([]interface{}{1.0, 2.0, 3.0})

	ao.Start()
	fo.Start()

	a.PortPushesAll([]interface{}{6.0, 20.0, 999.0, 10.0}, ao.Main().Out())
}

func TestBuiltinAggregate__PassMarkers(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := core.NewOperator(
		"test",
		func(op *core.Operator) {},
		nil,
		nil,
		nil,
		core.OperatorDef{
			ServiceDefs: map[string]*core.ServiceDef{
				"main": {
					In: core.TypeDef{Type: "trigger"},
					Out: core.TypeDef{Type: "map",
						Map: map[string]*core.TypeDef{
							"init":  {Type: "number"},
							"items": {Type: "stream", Stream: &core.TypeDef{Type: "number"}},
						},
					},
				},
			},
		},
	)
	r.NoError(err)
	a.NotNil(o)

	ao, err := MakeOperator(
		core.InstanceDef{
			Name:     "testOp",
			Operator: "slang.aggregate",
			Generics: map[string]*core.TypeDef{
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

	err = o.Main().Out().Connect(ao.Main().In())
	r.NoError(err)

	ao.Main().Out().Bufferize()
	ao.Delegate("iteration").Out().Bufferize()

	ao.Start()

	a.Equal(o.Main().Out().Map("items"), ao.Delegate("iteration").Out().StreamSource())

	o.Main().Out().Push(map[string]interface{}{"init": 0, "items": []interface{}{1, 2, 3}})

	i := ao.Delegate("iteration").Out().Stream().Map("item").Pull()
	a.True(ao.Delegate("iteration").Out().OwnBOS(i))
	a.True(o.Main().Out().Map("items").OwnBOS(i))
}
