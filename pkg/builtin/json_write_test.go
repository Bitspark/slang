package builtin

import (
	"testing"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/stretchr/testify/require"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/Bitspark/slang/pkg/utils"
)

func TestBuiltin_JsonWrite__CreatorFuncIsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFork := getBuiltinCfg("slang.encoding.json.write")
	a.NotNil(ocFork)
}

func TestBuiltin_JsonWrite__String(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.encoding.json.write",
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
	o.Main().In().Push("test")
	a.PortPushes(utils.Binary("\"test\""), o.Main().Out())
}
