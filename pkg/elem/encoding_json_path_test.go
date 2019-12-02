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
	a := assertions.New(t)

	ocFork := getBuiltinCfg(encodingJSONPathId)
	a.NotNil(ocFork)
}

func Test_JsonPath__String(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: encodingJSONPathId,
			Properties: map[string]interface{}{
				"paths": []interface{}{"name.last"},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()

	o.Main().In().Push(core.Binary(jsonDoc))
	a.PortPushes("Anderson", o.Main().Out().Map("name.last"))
	a.PortPushes(true, o.Main().Out().Map("valid"))
}

func Test_JsonPath__Invalid_Document(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: encodingJSONPathId,
			Properties: map[string]interface{}{
				"paths": []interface{}{"name.last"},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()
	o.Main().In().Push(core.Binary(`{"test"`))
	a.PortPushes(nil, o.Main().Out().Map("name.last"))
	a.PortPushes(false, o.Main().Out().Map("valid"))
}

func Test_JsonPath__NonExistent_Path(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: encodingJSONPathId,
			Properties: map[string]interface{}{
				"paths": []interface{}{"name.missing", "name.last"},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()
	o.Main().In().Push(core.Binary(jsonDoc))
	a.PortPushes("Anderson", o.Main().Out().Map("name.last"))
	a.PortPushes(nil, o.Main().Out().Map("name.missing"))
	a.PortPushes(true, o.Main().Out().Map("valid"))
}

func Test_JsonPath__Non_Primitive_Return(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: encodingJSONPathId,
			Properties: map[string]interface{}{
				"paths": []interface{}{"friends.#.last"},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()
	o.Main().In().Push(core.Binary(jsonDoc))
	a.PortPushes(`["Murphy","Craig","Fonder"]`, o.Main().Out().Map("friends.#.last"))
	a.PortPushes(true, o.Main().Out().Map("valid"))
}
