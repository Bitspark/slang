package elem

import (
	"github.com/stretchr/testify/require"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"testing"
)

func TestOperatorCSVRead__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocConst := getBuiltinCfg("77d60459-f8b5-4f4b-b293-740164c49a82")
	a.NotNil(ocConst)
}

func TestOperatorCSVRead__3Lines(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)
	co, err := buildOperator(
		core.InstanceDef{
			Operator: "77d60459-f8b5-4f4b-b293-740164c49a82",
			Generics: map[string]*core.TypeDef{
				"colMap": {
					Type: "map",
					Map: map[string]*core.TypeDef{
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
				"columns": []interface{}{"a", "b", "c"},
			},
		},
	)
	r.NoError(err)
	r.NotNil(co)

	co.Main().Out().Bufferize()
	co.Start()

	co.Main().In().Push("a,b,c\ne,f,g\nh,i,j")

	co.Main().Out().PullBOS()
	a.PortPushes(map[string]interface{}{"col_a": "e", "col_b": "f", "col_c": "g"}, co.Main().Out().Stream())
	a.PortPushes(map[string]interface{}{"col_a": "h", "col_b": "i", "col_c": "j"}, co.Main().Out().Stream())
	co.Main().Out().PullEOS()
}

func TestOperatorCSVRead__DifferentOrder(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)
	co, err := buildOperator(
		core.InstanceDef{
			Operator: "77d60459-f8b5-4f4b-b293-740164c49a82",
			Generics: map[string]*core.TypeDef{
				"colMap": {
					Type: "map",
					Map: map[string]*core.TypeDef{
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
				"columns": []interface{}{"a", "b", "c"},
			},
		},
	)
	r.NoError(err)
	r.NotNil(co)

	co.Main().Out().Bufferize()
	co.Start()

	co.Main().In().Push("b,c,a\ne,f,g\nh,i,j")

	co.Main().Out().PullBOS()
	a.PortPushes(map[string]interface{}{"col_b": "e", "col_c": "f", "col_a": "g"}, co.Main().Out().Stream())
	a.PortPushes(map[string]interface{}{"col_b": "h", "col_c": "i", "col_a": "j"}, co.Main().Out().Stream())
	co.Main().Out().PullEOS()
}

func TestOperatorCSVRead__NilCols(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)
	co, err := buildOperator(
		core.InstanceDef{
			Operator: "77d60459-f8b5-4f4b-b293-740164c49a82",
			Generics: map[string]*core.TypeDef{
				"colMap": {
					Type: "map",
					Map: map[string]*core.TypeDef{
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
				"columns": []interface{}{"a", "b", "c"},
			},
		},
	)
	r.NoError(err)
	r.NotNil(co)

	co.Main().Out().Bufferize()
	co.Start()

	co.Main().In().Push("x,b,c\ne,f,g\nh,i,j")

	co.Main().Out().PullBOS()
	a.PortPushes(map[string]interface{}{"col_a": nil, "col_b": "f", "col_c": "g"}, co.Main().Out().Stream())
	a.PortPushes(map[string]interface{}{"col_a": nil, "col_b": "i", "col_c": "j"}, co.Main().Out().Stream())
	co.Main().Out().PullEOS()
}
