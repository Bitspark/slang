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
				"path_names": []interface{}{"last_name"},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()
	data := map[string]interface{}{"last_name": "name.last", "document": core.Binary(jsonDoc)}
	o.Main().In().Push(data)
	a.PortPushes("Anderson", o.Main().Out().Map("last_name"))
	a.PortPushes(true, o.Main().Out().Map("valid"))
}

func Test_JsonPath__Invalid_Document(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: encodingJSONPathId,
			Properties: map[string]interface{}{
				"path_names": []interface{}{"last_name"},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()
	data := map[string]interface{}{"last_name": "name.last", "document": core.Binary(`{"test"`)}
	o.Main().In().Push(data)
	a.PortPushes(nil, o.Main().Out().Map("last_name"))
	a.PortPushes(false, o.Main().Out().Map("valid"))
}

func Test_JsonPath__NonExistent_Path(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: encodingJSONPathId,
			Properties: map[string]interface{}{
				"path_names": []interface{}{"name_missing", "name_last"},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()
	data := map[string]interface{}{"name_last": "name.last", "name_missing": "name.missing", "document": core.Binary(jsonDoc)}
	o.Main().In().Push(data)
	a.PortPushes("Anderson", o.Main().Out().Map("name_last"))
	a.PortPushes(nil, o.Main().Out().Map("name_missing"))
	a.PortPushes(true, o.Main().Out().Map("valid"))
}

func Test_JsonPath__Non_Primitive_Return(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: encodingJSONPathId,
			Properties: map[string]interface{}{
				"path_names": []interface{}{"friends_last"},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()
	data := map[string]interface{}{"friends_last": "friends.#.last", "document": core.Binary(jsonDoc)}
	o.Main().In().Push(data)
	a.PortPushes(`["Murphy","Craig","Fonder"]`, o.Main().Out().Map("friends_last"))
	a.PortPushes(true, o.Main().Out().Map("valid"))
}
