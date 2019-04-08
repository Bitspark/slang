package elem

import (
	"testing"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/stretchr/testify/require"
)

func Test_StreamMapAccess__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFork := getBuiltinCfg(streamMapAccessId)
	a.NotNil(ocFork)
}

func Test_StreamMapAccess__String(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: streamMapAccessId,
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
