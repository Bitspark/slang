package elem

import (
	"testing"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/stretchr/testify/require"
)

func Test_StreamWindow2__IsRegistered(t *testing.T) {
	Init()
	a := assertions.New(t)

	ocFork := getBuiltinCfg(streamWindow2Id)
	a.NotNil(ocFork)
}

func Test_StreamWindow2__Sliding1(t *testing.T) {
	Init()
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: streamWindow2Id,
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

	o.Main().In().Push("a")
	a.PortPushes(
		[]interface{}{"a"},
		o.Main().Out())

	o.Main().In().Push("b")
	a.PortPushes(
		[]interface{}{"a", "b"},
		o.Main().Out())

	o.Main().In().Push("c")
	a.PortPushes(
		[]interface{}{"a", "b", "c"},
		o.Main().Out())

	o.Main().In().Push("d")
	a.PortPushes(
		[]interface{}{"b", "c", "d"},
		o.Main().Out())

	o.Main().In().Push("e")
	a.PortPushes(
		[]interface{}{"c", "d", "e"},
		o.Main().Out())
}

/*
func Test_StreamWindow2__Sliding2(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: streamWindow2Id,
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

func Test_StreamWindow2__Sliding3(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: streamWindow2Id,
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

func Test_StreamWindow2__Jumping1(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: streamWindow2Id,
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

*/