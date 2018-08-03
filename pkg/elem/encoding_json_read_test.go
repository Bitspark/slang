package elem

import (
	"testing"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/stretchr/testify/require"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/Bitspark/slang/pkg/utils"
)

func Test_JsonRead__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFork := getBuiltinCfg("slang.encoding.JSONRead")
	a.NotNil(ocFork)
}

func Test_JsonRead__String(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.encoding.JSONRead",
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
	o.Main().In().Push(utils.Binary("\"test\""))
	a.PortPushes("test", o.Main().Out())
}

func Test_JsonRead__Complex(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.encoding.JSONRead",
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
	o.Main().In().Push(utils.Binary("{\"a\":[1,2,3],\"b\":true}"))
	a.PortPushes(map[string]interface{}{"a": []interface{}{1.0, 2.0, 3.0}, "b": true}, o.Main().Out())
}
