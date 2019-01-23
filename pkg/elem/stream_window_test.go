package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_StreamWindow__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFork := getBuiltinCfg("slang.stream.Window")
	a.NotNil(ocFork)
}

func Test_StreamWindow__Sliding1(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.stream.Window",
			Generics: map[string]*core.TypeDef{
				"itemType": {
					Type: "number",
				},
			},
			Properties: map[string]interface{}{
				"size":   3,
				"stride": 1,
				"fill":   false,
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()

	o.Main().In().Push([]interface{}{"a", "b", "c", "d", "e"})
	a.PortPushes([]interface{}{
		[]interface{}{"a"},
		[]interface{}{"a", "b"},
		[]interface{}{"a", "b", "c"},
		[]interface{}{"b", "c", "d"},
		[]interface{}{"c", "d", "e"},
	}, o.Main().Out())
}

func Test_StreamWindow__Sliding2(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.stream.Window",
			Generics: map[string]*core.TypeDef{
				"itemType": {
					Type: "number",
				},
			},
			Properties: map[string]interface{}{
				"size":   3,
				"stride": 2,
				"fill":   true,
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()

	o.Main().In().Push([]interface{}{"a", "b", "c", "d", "e"})
	a.PortPushes([]interface{}{
		[]interface{}{"a", "b", "c"},
		[]interface{}{"c", "d", "e"},
	}, o.Main().Out())
}

func Test_StreamWindow__Sliding3(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.stream.Window",
			Generics: map[string]*core.TypeDef{
				"itemType": {
					Type: "number",
				},
			},
			Properties: map[string]interface{}{
				"size":   3,
				"stride": 3,
				"fill":   true,
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()

	o.Main().In().Push([]interface{}{"a", "b", "c", "d", "e", "f"})
	a.PortPushes([]interface{}{
		[]interface{}{"a", "b", "c"},
		[]interface{}{"d", "e", "f"},
	}, o.Main().Out())
}

func Test_StreamWindow__Jumping1(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.stream.Window",
			Generics: map[string]*core.TypeDef{
				"itemType": {
					Type: "number",
				},
			},
			Properties: map[string]interface{}{
				"size":   2,
				"stride": 3,
				"fill":   true,
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()

	o.Main().In().Push([]interface{}{"a", "b", "c", "d", "e"})
	a.PortPushes([]interface{}{
		[]interface{}{"a", "b"},
		[]interface{}{"d", "e"},
	}, o.Main().Out())
}
