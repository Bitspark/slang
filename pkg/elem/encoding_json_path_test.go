package elem

import (
	"testing"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/stretchr/testify/require"
)

const jsonDoc = `
{
	"age":37,
	"children": ["Sara","Alex","Jack"],
	"fav.movie": "Deer Hunter",
	"friends": [
	  {"age": 44, "first": "Dale", "last": "Murphy"},
	  {"age": 68, "first": "Roger", "last": "Craig"},
	  {"age": 47, "first": "Jane", "last": "Fonder"}
	],
	"name": {"first": "Tom", "last": "Anderson"}
  }
`

func Test_JsonPath__IsRegistered(t *testing.T) {
	Init()
	a := assertions.New(t)

	ocFork := getBuiltinCfg(encodingJSONPathId)
	a.NotNil(ocFork)
}

func Test_JsonPath__String(t *testing.T) {
	Init()
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: encodingJSONPathId,
			Generics: core.Generics{
				"itemType": &core.TypeDef{
					Type: "string",
				},
			},
			Properties: map[string]interface{}{
				"path_names": []interface{}{
					map[string]interface{}{
						"name": "last_name",
						"query": "name.last",
					},
				},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()
	o.Main().In().Push(core.Binary(jsonDoc))
	a.PortPushes("Anderson", o.Main().Out().Map("last_name"))
	a.PortPushes(true, o.Main().Out().Map("valid"))
}

func Test_JsonPath__Number(t *testing.T) {
	Init()
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: encodingJSONPathId,
			Generics: core.Generics{
				"itemType": &core.TypeDef{
					Type: "number",
				},
			},
			Properties: map[string]interface{}{
				"path_names": []interface{}{
					map[string]interface{}{
						"name": "age",
						"query": "age",
					},
				},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()
	o.Main().In().Push(core.Binary(jsonDoc))
	a.PortPushes(37.0, o.Main().Out().Map("age"))
	a.PortPushes(true, o.Main().Out().Map("valid"))
}

func Test_JsonPath__Invalid_Document(t *testing.T) {
	Init()
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: encodingJSONPathId,
			Generics: core.Generics{
				"itemType": &core.TypeDef{
					Type: "string",
				},
			},
			Properties: map[string]interface{}{
				"path_names": []interface{}{
					map[string]interface{}{
						"name": "last_name",
						"query": "name.last",
					},
				},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()
	o.Main().In().Push(core.Binary("{test"))
	a.PortPushes(nil, o.Main().Out().Map("last_name"))
	a.PortPushes(false, o.Main().Out().Map("valid"))
}

func Test_JsonPath__NonExistent_Path(t *testing.T) {
	Init()
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: encodingJSONPathId,
			Generics: core.Generics{
				"itemType": &core.TypeDef{
					Type: "string",
				},
			},
			Properties: map[string]interface{}{
				"path_names": []interface{}{
					map[string]interface{}{
						"name": "last_name",
						"query": "name.last",
					},
					map[string]interface{}{
						"name": "name_missing",
						"query": "name.missing",
					},
				},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()
	o.Main().In().Push(core.Binary(jsonDoc))
	a.PortPushes("Anderson", o.Main().Out().Map("last_name"))
	a.PortPushes(nil, o.Main().Out().Map("name_missing"))
	a.PortPushes(true, o.Main().Out().Map("valid"))
}

func Test_JsonPath__Non_Primitive_Return(t *testing.T) {
	Init()
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: encodingJSONPathId,
			Generics: core.Generics{
				"itemType": &core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "string",
					},
				},
			},
			Properties: map[string]interface{}{
				"path_names": []interface{}{
					map[string]interface{}{
						"name": "friends_last",
						"query": "friends.#.last",
					},
				},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()
	o.Main().In().Push(core.Binary(jsonDoc))
	a.PortPushes([]interface {}{"Murphy", "Craig", "Fonder"}, o.Main().Out().Map("friends_last"))
	a.PortPushes(true, o.Main().Out().Map("valid"))
}

func Test_JsonPath__ParsePRTGHistoricDataPayload(t *testing.T) {
	Init()
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: encodingJSONPathId,
			Generics: core.Generics{
				"itemType": &core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "number",
					},
				},
			},
			Properties: map[string]interface{}{
				"path_names": []interface{}{
					map[string]interface{}{
						"name": "PV",
						"query": "histdata.#. Solar Energy Today ",
					},
				},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()
	o.Main().In().Push(core.Binary(`{
		"prtg-version": "23.1.82.2175",
		"treesize": 1200,
		"histdata": [
			{
			"datetime": "27.03.2023 00:00:00",
			" Power from PV ": "",
			" Solar Energy Today ": "",
			"coverage": "0 %"
			},
			{
			"datetime": "27.03.2023 12:37:00",
			" Power from PV ": 22.6910,
			" Solar Energy Today ": 112.5000,
			"coverage": "100 %"
			},
			{
			"datetime": "27.03.2023 12:38:00",
			" Power from PV ": 21.4000,
			" Solar Energy Today ": 112.9000,
			"coverage": "100 %"
			},
			{
			"datetime": "27.03.2023 12:39:00",
			" Power from PV ": 21.8630,
			" Solar Energy Today ": 113.2000,
			"coverage": "100 %"
			}
		]
	}`))
	a.PortPushes([]interface {}{"", 112.5, 112.9, 113.2}, o.Main().Out().Map("PV"))
	a.PortPushes(true, o.Main().Out().Map("valid"))
}