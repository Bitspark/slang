package elem

import (
	"testing"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/stretchr/testify/require"
)

func Test_TemplateFormat__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFork := getBuiltinCfg(stringTemplateId)
	a.NotNil(ocFork)
}

func Test_TemplateFormat__String(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: stringTemplateId,
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
