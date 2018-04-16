package builtin

import (
	"testing"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/stretchr/testify/require"
	"github.com/Bitspark/slang/tests/assertions"
)

func TestBuiltin_TemplateFormat__CreatorFuncIsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFork := getBuiltinCfg("slang.template.format")
	a.NotNil(ocFork)
}

func TestBuiltin_TemplateFormat__String(t *testing.T) {
	a := assertions.New(t)

	o, err := MakeOperator(
		core.InstanceDef{
			Operator: "slang.template.format",
			Properties: map[string]interface{}{
				"variables": []interface{}{"a"},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()
	o.Main().In().Push(map[string]interface{}{"a": "test", "content": []byte("__{a}__")})
	a.PortPushes([]byte("__test__"), o.Main().Out())
}
