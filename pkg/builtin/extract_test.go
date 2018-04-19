package builtin

import (
	"testing"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/stretchr/testify/require"
	"github.com/Bitspark/slang/tests/assertions"
)

func TestBuiltin_Extract__CreatorFuncIsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFork := getBuiltinCfg("slang.stream.extract")
	a.NotNil(ocFork)
}

func TestBuiltin_Extract__String(t *testing.T) {
	a := assertions.New(t)

	o, err := MakeOperator(
		core.InstanceDef{
			Operator: "slang.stream.extract",
			Generics: map[string]*core.TypeDef{
				"itemType": {
					Type: "string",
				},
			},
			Properties: core.Properties{
				"indexes": []interface{}{0, 2, 3},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()

	o.Main().In().Push([]interface{}{"test1", "test2", "test3"})
	a.PortPushes(map[string]interface{}{"el_0": "test1", "el_2": "test3", "el_3": nil}, o.Main().Out())

	o.Main().In().Push([]interface{}{})
	a.PortPushes(map[string]interface{}{"el_0": nil, "el_2": nil, "el_3": nil}, o.Main().Out())
}
