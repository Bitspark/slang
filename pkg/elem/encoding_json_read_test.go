package elem

import (
	"testing"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/stretchr/testify/require"
)

func Test_JsonRead__IsRegistered(t *testing.T) {
	Init()
	a := assertions.New(t)

	ocFork := getBuiltinCfg(encodingJSONReadId)
	a.NotNil(ocFork)
}

func Test_JsonRead__String(t *testing.T) {
	Init()
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: encodingJSONReadId,
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
	o.Main().In().Push(core.Binary("\"test\""))
	a.PortPushes("test", o.Main().Out().Map("item"))
	a.PortPushes(true, o.Main().Out().Map("valid"))
}

func Test_JsonRead__Invalid(t *testing.T) {
	Init()
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: encodingJSONReadId,
			Generics: map[string]*core.TypeDef{
				"itemType": {
					Type: "map",
					Map: map[string]*core.TypeDef{
						"a": {
							Type: "number",
						},
						"b": {
							Type: "boolean",
						},
					},
				},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()
	o.Main().In().Push(core.Binary("\"test\""))
	a.PortPushes(nil, o.Main().Out().Map("item").Map("a"))
	a.PortPushes(nil, o.Main().Out().Map("item").Map("b"))
	a.PortPushes(false, o.Main().Out().Map("valid"))
}

func Test_JsonRead__Complex(t *testing.T) {
	Init()
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: encodingJSONReadId,
			Generics: map[string]*core.TypeDef{
				"itemType": {
					Type: "map",
					Map: map[string]*core.TypeDef{
						"a": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type: "number",
							},
						},
						"b": {
							Type: "boolean",
						},
					},
				},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()
	o.Main().In().Push(core.Binary("{\"a\":[1,2,3],\"b\":true}"))
	a.PortPushes(map[string]interface{}{"a": []interface{}{1.0, 2.0, 3.0}, "b": true}, o.Main().Out().Map("item"))
	a.PortPushes(true, o.Main().Out().Map("valid"))
}

func Test_JsonRead__Subsets(t *testing.T) {
	itemType := &core.TypeDef{Type: "map", Map: core.TypeDefMap{
		"name": {Type: "string"},
		"details": {Type: "map", Map: core.TypeDefMap{
			"active": {Type: "boolean"},
		}},
		"rows": {Type: "stream", Stream: &core.TypeDef{Type: "map", Map: core.TypeDefMap{
			"value": {Type: "number"},
		}}},
	}}
	for _, tc := range []struct {
		name  string
		input string
		valid bool
	}{
		{"extra fields at every level", `{"name":"Ada","ignored":42,"details":{"active":true,"extra":"ok"},"rows":[{"value":1,"extra":2},{"value":3,"extra":4}]}`, true},
		{"missing required field", `{"details":{"active":true},"rows":[]}`, false},
		{"missing nested field", `{"name":"Ada","details":{"extra":true},"rows":[]}`, false},
		{"wrong nested type", `{"name":"Ada","details":{"active":"yes"},"rows":[]}`, false},
		{"wrong stream item type", `{"name":"Ada","details":{"active":true},"rows":[{"value":"one"}]}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			Init()
			o, err := buildOperator(core.InstanceDef{
				Operator: encodingJSONReadId,
				Generics: core.Generics{"itemType": itemType},
			})
			require.NoError(t, err)
			o.Main().Out().Bufferize()
			o.Start()
			defer o.Stop()
			o.Main().In().Push(core.Binary(tc.input))
			require.Equal(t, tc.valid, o.Main().Out().Map("valid").Pull())
			if tc.valid {
				require.Equal(t, map[string]interface{}{
					"name":    "Ada",
					"details": map[string]interface{}{"active": true},
					"rows": []interface{}{
						map[string]interface{}{"value": float64(1)},
						map[string]interface{}{"value": float64(3)},
					},
				}, o.Main().Out().Map("item").Pull())
			}
		})
	}
}
