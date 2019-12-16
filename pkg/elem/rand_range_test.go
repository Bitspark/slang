package elem

import (
	"testing"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/stretchr/testify/require"
)

func buildRandOperator(t *testing.T) *core.Operator {
	o, err := buildOperator(
		core.InstanceDef{
			Operator: randRangeId,
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()
	return o
}

func Test_Rand_Range__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFork := getBuiltinCfg(randRangeId)
	a.NotNil(ocFork)
}

func Test_Rand_Range(t *testing.T) {
	a := assertions.New(t)

	o := buildRandOperator(t)
	o.Main().In().Push(map[string]interface{}{"min": 1, "max": 1})
	a.PortPushes(1, o.Main().Out())

	o = buildRandOperator(t)
	o.Main().In().Push(map[string]interface{}{"min": 0, "max": 0})
	a.PortPushes(0, o.Main().Out())

	// this breaks type conversion
	o = buildRandOperator(t)
	o.Main().In().Push(map[string]interface{}{"min": 0.0, "max": 0})
	a.PortPushes(nil, o.Main().Out())

	// this breaks type conversion
	o = buildRandOperator(t)
	o.Main().In().Push(map[string]interface{}{"min": "0.0", "max": 0})
	a.PortPushes(nil, o.Main().Out())
}
func Test_Rand_Range_Negative_Values(t *testing.T) {
	a := assertions.New(t)

	o := buildRandOperator(t)

	o.Main().In().Push(map[string]interface{}{"min": -2, "max": -2})
	a.PortPushes(-2, o.Main().Out())
}
