package builtin

import (
	"testing"
	"slang/tests/assertions"
	"slang/core"
	"github.com/stretchr/testify/require"
)

func TestOperatorConst__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocConst := getBuiltinCfg("const")
	a.NotNil(ocConst)
}

func TestOperatorConst__NoProps(t *testing.T) {
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

func TestBuiltinConst__NoGenerics(t *testing.T) {
	a := assertions.New(t)
	ao, err := MakeOperator(
		core.InstanceDef{
			Operator:   "const",
			Properties: core.Properties{"value": 1.0},
		},
	)
	a.Error(err)
	a.Nil(ao)
}

func TestBuiltinConst__Correct(t *testing.T) {
	a := assertions.New(t)
	ao, err := MakeOperator(
		core.InstanceDef{
			Operator: "const",
			Generics: map[string]*core.PortDef{
				"valueType": {
					Type: "number",
				},
			},
			Properties: core.Properties{"value": 1.0},
		},
	)
	require.NoError(t, err)
	a.NotNil(ao)

	a.Equal(core.TYPE_TRIGGER, ao.In().Type())
	a.Equal(core.TYPE_NUMBER, ao.Out().Type())
}

func TestBuiltinConst__PushBoolean(t *testing.T) {
	a := assertions.New(t)
	ao, err := MakeOperator(
		core.InstanceDef{
			Operator: "const",
			Generics: map[string]*core.PortDef{
				"valueType": {
					Type: "boolean",
				},
			},
			Properties: core.Properties{"value": true},
		},
	)
	require.NoError(t, err)
	a.NotNil(ao)

	a.Equal(core.TYPE_BOOLEAN, ao.Out().Type())

	ao.Out().Bufferize()
	ao.Start()

	for i := 0; i < 20; i++ {
		ao.In().Push(1)
	}

	a.PortPushes([]interface{}{true, true, true, true, true, true, true, true, true, true}, ao.Out())
	a.PortPushes([]interface{}{true, true, true, true, true, true, true, true, true, true}, ao.Out())
	// ...
}

func TestBuiltinConst__PushStream(t *testing.T) {
	a := assertions.New(t)
	ao, err := MakeOperator(
		core.InstanceDef{
			Operator: "const",
			Generics: map[string]*core.PortDef{
				"valueType": {
					Type: "stream",
					Stream: &core.PortDef{
						Type: "primitive",
					},
				},
			},
			Properties: core.Properties{"value": []interface{}{1.0, "slang", true}},
		},
	)
	require.NoError(t, err)
	a.NotNil(ao)

	a.Equal(core.TYPE_STREAM, ao.Out().Type())

	ao.Out().Bufferize()
	ao.Start()

	for i := 0; i < 3; i++ {
		ao.In().Push(1)
	}

	a.PortPushes([]interface{}{[]interface{}{1.0, "slang", true}}, ao.Out())
	a.PortPushes([]interface{}{[]interface{}{1.0, "slang", true}}, ao.Out())
	a.PortPushes([]interface{}{[]interface{}{1.0, "slang", true}}, ao.Out())
	// ...
}

func TestBuiltinConst__PushMap(t *testing.T) {
	a := assertions.New(t)
	ao, err := MakeOperator(
		core.InstanceDef{
			Operator: "const",
			Generics: map[string]*core.PortDef{
				"valueType": {
					Type: "map",
					Map: map[string]*core.PortDef{
						"a": {
							Type: "number",
						},
						"b": {
							Type: "primitive",
						},
					},
				},
			},
			Properties: core.Properties{"value": map[string]interface{}{"a": 1.0, "b": false}},
		},
	)
	require.NoError(t, err)
	a.NotNil(ao)

	a.Equal(core.TYPE_MAP, ao.Out().Type())

	ao.Out().Bufferize()
	ao.Start()

	for i := 0; i < 3; i++ {
		ao.In().Push(1)
	}

	a.PortPushes([]interface{}{map[string]interface{}{"a": 1.0, "b": false}}, ao.Out())
	a.PortPushes([]interface{}{map[string]interface{}{"a": 1.0, "b": false}}, ao.Out())
	a.PortPushes([]interface{}{map[string]interface{}{"a": 1.0, "b": false}}, ao.Out())
	// ...
}

func TestOperatorConst__SimpleNumber(t *testing.T) {
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

func TestOperatorConst__ComplexStreamMap(t *testing.T) {
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

func TestOperatorConst__PassMarkers(t *testing.T) {
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