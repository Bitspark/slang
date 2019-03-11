package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_GoogleAnalytics(t *testing.T) {
	r := require.New(t)

	ao, err := buildOperator(core.InstanceDef{
		Operator: "257b55e8-b27f-4ef2-8581-5ef3f0af4491",
		Properties: map[string]interface{}{
			"jsonFile": "D:/Development/serious-energy-234221-45131bb46a9d.json",
			"gaid":     "ga:111009352",
		},
	})
	r.NoError(err)
	r.NotNil(ao)

	ao.Main().Out().Bufferize()
	ao.Start()

	ao.Main().In().Push(nil)
	userNumber := ao.Main().Out().Pull()

	r.Equal(0.0, userNumber)
}
