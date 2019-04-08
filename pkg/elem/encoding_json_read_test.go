package elem

import (
	"testing"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/stretchr/testify/require"
)

func Test_JsonRead__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFork := getBuiltinCfg(encodingJSONReadId)
	a.NotNil(ocFork)
}

func Test_JsonRead__String(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: encodingJSONReadId,
			Generics: map[string]*core.TypeDef{
				"itemType": {
					Type: "string",
				},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()
	o.Main().In().Push(core.Binary("\"test\""))
	a.PortPushes("test", o.Main().Out().Map("item"))
	a.PortPushes(true, o.Main().Out().Map("valid"))
}

func Test_JsonRead__Invalid(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: encodingJSONReadId,
			Generics: map[string]*core.TypeDef{
				"itemType": {
					Type: "map",
					Map: map[string]*core.TypeDef{
						"a": {
							Type: "number",
						},
						"b": {
							Type: "boolean",
						},
					},
				},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()
	o.Main().In().Push(core.Binary("\"test\""))
	a.PortPushes(nil, o.Main().Out().Map("item").Map("a"))
	a.PortPushes(nil, o.Main().Out().Map("item").Map("b"))
	a.PortPushes(false, o.Main().Out().Map("valid"))
}

func Test_JsonRead__Complex(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: encodingJSONReadId,
			Generics: map[string]*core.TypeDef{
				"itemType": {
					Type: "map",
					Map: map[string]*core.TypeDef{
						"a": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type: "number",
							},
						},
						"b": {
							Type: "boolean",
						},
					},
				},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()
	o.Main().In().Push(core.Binary("{\"a\":[1,2,3],\"b\":true}"))
	a.PortPushes(map[string]interface{}{"a": []interface{}{1.0, 2.0, 3.0}, "b": true}, o.Main().Out().Map("item"))
	a.PortPushes(true, o.Main().Out().Map("valid"))
}
