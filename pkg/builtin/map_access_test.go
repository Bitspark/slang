package builtin

import (
	"testing"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/stretchr/testify/require"
	"github.com/Bitspark/slang/tests/assertions"
)

func TestBuiltin_MapAccess__CreatorFuncIsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFork := getBuiltinCfg("slang.stream.mapAccess")
	a.NotNil(ocFork)
}

func TestBuiltin_MapAccess__String(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.stream.mapAccess",
			Generics: map[string]*core.TypeDef{
				"keyType": {
					Type: "string",
				},
				"valueType": {
					Type: "number",
				},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()

	o.Main().In().Push(map[string]interface{}{"key": "a", "stream": []interface{}{map[string]interface{}{"key": "a", "value": 1}}})
	a.PortPushes(1, o.Main().Out())

	o.Main().In().Push(map[string]interface{}{"key": "c", "stream": []interface{}{map[string]interface{}{"key": "a", "value": 1}}})
	a.PortPushes(nil, o.Main().Out())

	o.Main().In().Push(map[string]interface{}{"key": "b", "stream": []interface{}{map[string]interface{}{"key": "a", "value": 1}, map[string]interface{}{"key": "b", "value": 2}}})
	a.PortPushes(2, o.Main().Out())
}
