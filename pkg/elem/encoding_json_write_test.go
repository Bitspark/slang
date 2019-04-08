package elem

import (
	"testing"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/stretchr/testify/require"
)

func Test_JsonWrite__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFork := getBuiltinCfg(encodingJSONWriteId)
	a.NotNil(ocFork)
}

func Test_JsonWrite__String(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: encodingJSONWriteId,
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
	a.PortPushes(core.Binary("\"test\""), o.Main().Out())
}
