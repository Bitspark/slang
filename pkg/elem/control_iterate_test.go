package elem

import (
	"testing"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/stretchr/testify/require"
)

func Test_CtrlIterate__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocAgg := getBuiltinCfg(controlIterateId)
	a.NotNil(ocAgg)
}

func Test_CtrlIterate__PassOtherMarkers(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	ao, err := buildOperator(core.InstanceDef{
		Operator: controlIterateId,
		Generics: map[string]*core.TypeDef{
			"stateType": {
				Type: "number",
			},
			"inItemType": {
				Type: "number",
			},
			"outItemType": {
				Type: "number",
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, ao)

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
						Stream: &core.TypeDef{Type: "map",
							Map: map[string]*core.TypeDef{
								"result": {Type: "number"},
								"items":  {Type: "stream", Stream: &core.TypeDef{Type: "number"}}}}},
				},
			},
		},
	)
	require.NoError(t, err)
	require.NotNil(t, do)

	ao.SetParent(do)

	r.NoError(do.Main().In().Stream().Map("init").Connect(ao.Main().In().Map("initial")))
	r.NoError(do.Main().In().Stream().Map("items").Connect(ao.Main().In().Map("items")))
	r.NoError(ao.Delegate("iterator").Out().Connect(ao.Delegate("iterator").In()))
	r.NoError(ao.Main().Out().Connect(do.Main().Out().Stream()))

	do.Main().Out().Bufferize()

	do.Start()

	do.Main().In().Push([]interface{}{map[string]interface{}{"init": 0.0, "items": []interface{}{}}})
	a.PortPushesAll([]interface{}{[]interface{}{map[string]interface{}{"result": 0.0, "items": []interface{}{}}}}, do.Main().Out())
}

func Test_CtrlIterate__SimpleAggregation(t *testing.T) {
	a := assertions.New(t)
	ao, err := buildOperator(
		core.InstanceDef{
			Operator: controlIterateId,
			Generics: map[string]*core.TypeDef{
				"inItemType": {
					Type: "number",
				},
				"outItemType": {
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
					out.Map("state").Push(m["state"].(float64) + m["item"].(float64))
					out.Map("item").Push(-m["state"].(float64))
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
					Out: core.TypeDef{Type: "map", Map: map[string]*core.TypeDef{"state": {Type: "number"}, "item": {Type: "number"}}},
				},
			},
		},
	)
	require.NoError(t, err)

	// Connect
	require.NoError(t, ao.Delegate("iterator").Out().Connect(fo.Main().In()))
	require.NoError(t, fo.Main().Out().Connect(ao.Delegate("iterator").In()))

	ao.Main().Out().Bufferize()

	ao.Main().In().Map("initial").Push(0.0)
	ao.Main().In().Map("items").Push([]interface{}{1.0, 2.0, 3.0})

	ao.Main().In().Map("initial").Push(8.0)
	ao.Main().In().Map("items").Push([]interface{}{2.0, 4.0, 6.0})

	ao.Main().In().Map("initial").Push(999.0)
	ao.Main().In().Map("items").Push([]interface{}{})

	ao.Main().In().Map("initial").Push(4.0)
	ao.Main().In().Map("items").Push([]interface{}{1.0, 2.0, 3.0})

	ao.Start()
	fo.Start()

	a.PortPushesAll([]interface{}{
		map[string]interface{}{"result": 6.0, "items": []interface{}{0.0, -1.0, -3.0}},
		map[string]interface{}{"result": 20.0, "items": []interface{}{-8.0, -10.0, -14.0}},
		map[string]interface{}{"result": 999.0, "items": []interface{}{}},
		map[string]interface{}{"result": 10.0, "items": []interface{}{-4.0, -5.0, -7.0}},
	}, ao.Main().Out())
}
