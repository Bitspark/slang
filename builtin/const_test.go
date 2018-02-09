package builtin

import (
	"testing"
	"slang/tests/assertions"
	"slang/core"
	"github.com/stretchr/testify/require"
)

func TestOperatorCreator__Const__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocConst := getBuiltinCfg("const")
	a.NotNil(ocConst)
}

func TestOperatorCreator__Const__RequireValue(t *testing.T) {
	a := assertions.New(t)
	co, err := MakeOperator(
		core.InstanceDef{
			Operator: "const",
			Generics: map[string]*core.PortDef{
				"valueType": {
					Type: "number",
				},
			},
		},
	)
	a.Error(err)
	a.Nil(co)
}

func TestOperatorCreator__Const__SimpleNumber(t *testing.T) {
	a := assertions.New(t)
	co, err := MakeOperator(
		core.InstanceDef{
			Operator: "const",
			Generics: map[string]*core.PortDef{
				"valueType": {
					Type: "number",
				},
			},
			Properties: map[string]interface{}{
				"value": 5.0,
			},
		},
	)
	require.NoError(t, err)
	a.NotNil(co)

	co.Out().Bufferize()
	co.Start()

	co.In().Push(true)
	a.PortPushes([]interface{}{5.0}, co.Out())
}

func TestOperatorCreator__Const__ComplexStreamMap(t *testing.T) {
	a := assertions.New(t)
	co, err := MakeOperator(
		core.InstanceDef{
			Operator: "const",
			Generics: map[string]*core.PortDef{
				"valueType": {
					Type: "map",
					Map: map[string]*core.PortDef{
						"a": {
							Type: "stream",
							Stream: &core.PortDef{
								Type: "number",
							},
						},
						"b": {
							Type: "string",
						},
					},
				},
			},
			Properties: map[string]interface{}{
				"value": map[string]interface{}{"a": []interface{}{1.0, 2.0, 3.0}, "b": "test"},
			},
		},
	)
	require.NoError(t, err)
	a.NotNil(co)

	co.Out().Map("a").Stream().Bufferize()
	co.Out().Map("b").Bufferize()
	co.Start()

	co.In().Push(true)
	a.PortPushes([]interface{}{[]interface{}{1.0, 2.0, 3.0}}, co.Out().Map("a"))
	a.PortPushes([]interface{}{"test"}, co.Out().Map("b"))
}

func TestOperatorCreator__Const__PassMarkers(t *testing.T) {
	a := assertions.New(t)
	co, err := MakeOperator(
		core.InstanceDef{
			Operator: "const",
			Generics: map[string]*core.PortDef{
				"valueType": {
					Type: "number",
				},
			},
			Properties: map[string]interface{}{
				"value": 5.0,
			},
		},
	)
	require.NoError(t, err)
	a.NotNil(co)

	co.Out().Bufferize()
	co.Start()

	bos := core.BOS{}
	eos := core.BOS{}

	co.In().Push(bos)
	co.In().Push(true)
	co.In().Push(eos)

	a.PortPushes([]interface{}{bos, 5.0, eos}, co.Out())
}
