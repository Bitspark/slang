package builtin

import (
	"testing"
	"slang/tests/assertions"
	"slang/core"
)

func TestOperatorCreator__Agg__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocAgg := getBuiltinCfg("agg")
	a.NotNil(ocAgg)
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
	a.NoError(err)
	a.NotNil(ao)

	// Add function operator
	fo, _ := core.NewOperator("add", func(in, out *core.Port, store interface{}) {
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

	// Connect
	a.NoError(ao.Out().Map("iteration").Stream().Connect(fo.In()))
	a.NoError(fo.Out().Connect(ao.In().Map("state").Stream()))

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
