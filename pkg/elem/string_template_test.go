package elem

import (
	"testing"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/stretchr/testify/require"
	"github.com/Bitspark/slang/tests/assertions"
)

func Test_TemplateFormat__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFork := getBuiltinCfg("slang.string.Template")
	a.NotNil(ocFork)
}

func Test_TemplateFormat__String(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.string.Template",
			Properties: map[string]interface{}{
				"variables": []interface{}{"a"},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()
	o.Main().In().Push(map[string]interface{}{"a": "test", "content": "__{a}__"})
	a.PortPushes("__test__", o.Main().Out())
}
