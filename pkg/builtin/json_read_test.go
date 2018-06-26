package builtin

import (
	"testing"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/stretchr/testify/require"
	"github.com/Bitspark/slang/tests/assertions"
)

func TestBuiltin_JsonRead__CreatorFuncIsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFork := getBuiltinCfg("slang.encoding.json.read")
	a.NotNil(ocFork)
}

func TestBuiltin_JsonRead__String(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.encoding.json.read",
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
	o.Main().In().Push([]byte("\"test\""))
	a.PortPushes("test", o.Main().Out())
}

func TestBuiltin_JsonRead__Complex(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.encoding.json.read",
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
	o.Main().In().Push([]byte("{\"a\":[1,2,3],\"b\":true}"))
	a.PortPushes(map[string]interface{}{"a": []interface{}{1.0, 2.0, 3.0}, "b": true}, o.Main().Out())
}
