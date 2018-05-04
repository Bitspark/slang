package builtin

import (
	"github.com/stretchr/testify/require"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"testing"
)

func TestOperatorCSVRead__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocConst := getBuiltinCfg("slang.encoding.csv.read")
	a.NotNil(ocConst)
}

func TestOperatorCSVRead__3Lines(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)
	co, err := MakeOperator(
		core.InstanceDef{
			Operator: "slang.encoding.csv.read",
			Generics: map[string]*core.PortDef{
				"colMap": {
					Type: "map",
					Map: map[string]*core.PortDef{
						"a": {
							Type: "string",
						},
						"b": {
							Type: "string",
						},
						"c": {
							Type: "string",
						},
					},
				},
			},
			Properties: map[string]interface{}{
				"delimiter": ",",
			},
		},
	)
	r.NoError(err)
	r.NotNil(co)

	co.Out().Bufferize()
	co.Start()

	co.In().Push("a,b,c\ne,f,g\nh,i,j")

	co.Out().PullBOS()
	a.PortPushes(map[string]interface{}{"a": "e", "b": "f", "c": "g"}, co.Out().Stream())
	a.PortPushes(map[string]interface{}{"a": "h", "b": "i", "c": "j"}, co.Out().Stream())
	co.Out().PullEOS()
}
