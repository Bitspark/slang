package elem

import (
	"github.com/stretchr/testify/require"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"testing"
)

func Test_DataConstant__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocConst := getBuiltinCfg("slang.data.Value")
	a.NotNil(ocConst)
}

func Test_DataConstant__NoProps(t *testing.T) {
	a := assertions.New(t)
	co, err := MakeOperator(
		core.InstanceDef{
			Operator: "const",
			Generics: map[string]*core.TypeDef{
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
	ao, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.data.Value",
			Generics: map[string]*core.TypeDef{
				"valueType": {
					Type: "number",
				},
			},
			Properties: core.Properties{"value": 1.0},
		},
	)
	require.NoError(t, err)
	a.NotNil(ao)

	a.Equal(core.TYPE_TRIGGER, ao.Main().In().Type())
	a.Equal(core.TYPE_NUMBER, ao.Main().Out().Type())
}

func TestBuiltinConst__PushBoolean(t *testing.T) {
	a := assertions.New(t)
	ao, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.data.Value",
			Generics: map[string]*core.TypeDef{
				"valueType": {
					Type: "boolean",
				},
			},
			Properties: core.Properties{"value": true},
		},
	)
	require.NoError(t, err)
	a.NotNil(ao)

	a.Equal(core.TYPE_BOOLEAN, ao.Main().Out().Type())

	ao.Main().Out().Bufferize()
	ao.Start()

	for i := 0; i < 20; i++ {
		ao.Main().In().Push(1)
	}

	a.PortPushesAll([]interface{}{true, true, true, true, true, true, true, true, true, true}, ao.Main().Out())
	a.PortPushesAll([]interface{}{true, true, true, true, true, true, true, true, true, true}, ao.Main().Out())
	// ...
}

func TestBuiltinConst__PushStream(t *testing.T) {
	a := assertions.New(t)
	ao, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.data.Value",
			Generics: map[string]*core.TypeDef{
				"valueType": {
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "primitive",
					},
				},
			},
			Properties: core.Properties{"value": []interface{}{1.0, "slang", true}},
		},
	)
	require.NoError(t, err)
	a.NotNil(ao)

	a.Equal(core.TYPE_STREAM, ao.Main().Out().Type())

	ao.Main().Out().Bufferize()
	ao.Start()

	for i := 0; i < 3; i++ {
		ao.Main().In().Push(1)
	}

	a.PortPushesAll([]interface{}{[]interface{}{1.0, "slang", true}}, ao.Main().Out())
	a.PortPushesAll([]interface{}{[]interface{}{1.0, "slang", true}}, ao.Main().Out())
	a.PortPushesAll([]interface{}{[]interface{}{1.0, "slang", true}}, ao.Main().Out())
	// ...
}

func TestBuiltinConst__PushMap(t *testing.T) {
	a := assertions.New(t)
	ao, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.data.Value",
			Generics: map[string]*core.TypeDef{
				"valueType": {
					Type: "map",
					Map: map[string]*core.TypeDef{
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

	a.Equal(core.TYPE_MAP, ao.Main().Out().Type())

	ao.Main().Out().Bufferize()
	ao.Start()

	for i := 0; i < 3; i++ {
		ao.Main().In().Push(1)
	}

	a.PortPushesAll([]interface{}{map[string]interface{}{"a": 1.0, "b": false}}, ao.Main().Out())
	a.PortPushesAll([]interface{}{map[string]interface{}{"a": 1.0, "b": false}}, ao.Main().Out())
	a.PortPushesAll([]interface{}{map[string]interface{}{"a": 1.0, "b": false}}, ao.Main().Out())
	// ...
}

func Test_DataConstant__SimpleNumber(t *testing.T) {
	a := assertions.New(t)
	co, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.data.Value",
			Generics: map[string]*core.TypeDef{
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

	co.Main().Out().Bufferize()
	co.Start()

	co.Main().In().Push(true)
	a.PortPushesAll([]interface{}{5.0}, co.Main().Out())
}

func Test_DataConstant__ComplexStreamMap(t *testing.T) {
	a := assertions.New(t)
	co, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.data.Value",
			Generics: map[string]*core.TypeDef{
				"valueType": {
					Type: "map",
					Map: map[string]*core.TypeDef{
						"a": {
							Type: "stream",
							Stream: &core.TypeDef{
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

	co.Main().Out().Bufferize()
	co.Start()

	co.Main().In().Push(true)
	a.PortPushesAll([]interface{}{[]interface{}{1.0, 2.0, 3.0}}, co.Main().Out().Map("a"))
	a.PortPushesAll([]interface{}{"test"}, co.Main().Out().Map("b"))
}

func Test_DataConstant__PassMarkers(t *testing.T) {
	a := assertions.New(t)
	co, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.data.Value",
			Generics: map[string]*core.TypeDef{
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

	co.Main().Out().Bufferize()
	co.Start()

	bos := core.BOS{}
	eos := core.BOS{}

	co.Main().In().Push(bos)
	co.Main().In().Push(true)
	co.Main().In().Push(eos)

	a.PortPushesAll([]interface{}{bos, 5.0, eos}, co.Main().Out())
}
