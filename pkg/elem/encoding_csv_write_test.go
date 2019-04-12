package elem

import (
	"testing"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/stretchr/testify/require"
)

func TestOperatorCSVWrite__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocConst := getBuiltinCfg(encodingCSVWriteId)
	a.NotNil(ocConst)
}

func TestOperatorCSVWrite__3Lines(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)
	co, err := buildOperator(
		core.InstanceDef{
			Operator: encodingCSVWriteId,
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
				"delimiter":     ",",
				"includeHeader": false,
				"columns":       []interface{}{"a", "b", "c"},
			},
		},
	)
	r.NoError(err)
	r.NotNil(co)

	co.Main().Out().Bufferize()
	co.Start()

	co.Main().In().Push([]interface{}{map[string]interface{}{"col_a": "e", "col_b": "f", "col_c": "g"}})

	a.PortPushes("e,f,g\n", co.Main().Out())
}

func TestOperatorCSVWrite__DifferentOrder(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)
	co, err := buildOperator(
		core.InstanceDef{
			Operator: encodingCSVWriteId,
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
				"delimiter":     ",",
				"includeHeader": false,
				"columns":       []interface{}{"a", "b", "c"},
			},
		},
	)
	r.NoError(err)
	r.NotNil(co)

	co.Main().Out().Bufferize()
	co.Start()

	co.Main().In().Push([]interface{}{
		map[string]interface{}{"col_b": "e", "col_c": "f", "col_a": "g"},
		map[string]interface{}{"col_c": "h", "col_b": "i", "col_a": "j"}})

	a.PortPushes("g,e,f\nj,i,h\n", co.Main().Out())
}
func TestOperatorCSVWrite__IncludeHeader(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)
	co, err := buildOperator(
		core.InstanceDef{
			Operator: encodingCSVWriteId,
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
				"delimiter":     ",",
				"includeHeader": true,
				"columns":       []interface{}{"a", "b", "c"},
			},
		},
	)
	r.NoError(err)
	r.NotNil(co)

	co.Main().Out().Bufferize()
	co.Start()

	co.Main().In().Push([]interface{}{
		map[string]interface{}{"col_c": "h", "col_b": "i", "col_a": "j"}})

	a.PortPushes("a,b,c\nj,i,h\n", co.Main().Out())
}
