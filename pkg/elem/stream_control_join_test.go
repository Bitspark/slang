package elem

import (
	"testing"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_StreamJoin__IsRegistered(t *testing.T) {
	Init()
	a := assertions.New(t)

	ocFork := getBuiltinCfg(uuid.MustParse("fb174c53-80bd-4e29-955a-aafe33ebfb31"))
	a.NotNil(ocFork)
}

func Test_StreamJoin__BestCase_CompleteStreams(t *testing.T) {
	Init()
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: streamCtrlJoinCfg.blueprint.Id,
			Properties: core.Properties{
				"streams": []interface{}{"a", "b", "c"},
			},
			Generics: core.Generics{
				"itemType": {Type: "string"},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()

	o.Main().In().Map("stream_a").Push([]interface{}{"a1", "a2", "a3", "a4", "a5"})
	o.Main().In().Map("stream_b").Push([]interface{}{"b1", "b2", "b3", "b4", "b5"})
	o.Main().In().Map("stream_c").Push([]interface{}{"c1", "c2", "c3", "c4", "c5"})

	a.PortPushes([]interface {}{
		"a1", "b1", "c1",
		"a2", "b2", "c2",
		"a3", "b3", "c3",
		"a4", "b4", "c4",
		"a5", "b5", "c5",
	}, o.Main().Out())
}

func Test_StreamJoin__MapInput_CompleteStreams(t *testing.T) {
	Init()
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: streamCtrlJoinCfg.blueprint.Id,
			Properties: core.Properties{
				"streams": []interface{}{"a", "b"},
			},
			Generics: core.Generics{
				"itemType": {Type: "map", Map: core.TypeDefMap{"K": {Type: "string"}, "V": {Type: "number"}}},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()

	o.Main().In().Map("stream_a").Push([]interface{}{
		map[string]interface{}{"K": "a", "V": 1},
		map[string]interface{}{"K": "a", "V": 2},
		map[string]interface{}{"K": "a", "V": 3},
		map[string]interface{}{"K": "a", "V": 4},
	})
	o.Main().In().Map("stream_b").Push([]interface{}{
		map[string]interface{}{"K": "b", "V": 1},
		map[string]interface{}{"K": "b", "V": 2},
	})

	a.PortPushes([]interface{}{
		map[string]interface{}{"K": "a", "V": 1},
		map[string]interface{}{"K": "b", "V": 1},
		map[string]interface{}{"K": "a", "V": 2},
		map[string]interface{}{"K": "b", "V": 2},
		map[string]interface{}{"K": "a", "V": 3},
		map[string]interface{}{"K": "a", "V": 4},
	}, o.Main().Out())
}

func Test_StreamJoin__BestCase_IncompleteStreams(t *testing.T) {
	Init()
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: streamCtrlJoinCfg.blueprint.Id,
			Properties: core.Properties{
				"streams": []interface{}{"a", "b", "c"},
			},
			Generics: core.Generics{
				"itemType": {Type: "string"},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()

	o.Main().In().Map("stream_a").PushBOS()
	o.Main().In().Map("stream_b").PushBOS()
	o.Main().In().Map("stream_c").PushBOS()

	o.Main().In().Map("stream_a").Stream().Push("a1")
	o.Main().In().Map("stream_b").Stream().Push("b1")
	o.Main().In().Map("stream_c").Stream().Push("c1")

	a.True(core.IsBOS(o.Main().Out().Stream().Pull()))
	a.PortPushes("a1", o.Main().Out().Stream())
	a.PortPushes("b1", o.Main().Out().Stream())
	a.PortPushes("c1", o.Main().Out().Stream())

	o.Main().In().Map("stream_a").Stream().Push("a2")
	o.Main().In().Map("stream_b").Stream().Push("b2")
	o.Main().In().Map("stream_c").Stream().Push("c2")

	a.PortPushes("a2", o.Main().Out().Stream())
	a.PortPushes("b2", o.Main().Out().Stream())
	a.PortPushes("c2", o.Main().Out().Stream())

	o.Main().In().Map("stream_a").Stream().Push("a3")
	o.Main().In().Map("stream_c").Stream().Push("c3")

	a.PortPushes("a3", o.Main().Out().Stream())
	a.PortPushes("c3", o.Main().Out().Stream())
}

func Test_StreamJoin__Uneven_CompleteStreams(t *testing.T) {
	Init()
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: streamCtrlJoinCfg.blueprint.Id,
			Properties: core.Properties{
				"streams": []interface{}{"a", "b", "c"},
			},
			Generics: core.Generics{
				"itemType": {Type: "string"},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()


	o.Main().In().Map("stream_a").Push([]interface{}{"a1", "a2", "a3"})
	o.Main().In().Map("stream_b").Push([]interface{}{"b1", "b2", "b3", "b4", "b5"})
	o.Main().In().Map("stream_c").Push([]interface{}{"c1", "c2", "c3", "c4"})

	a.PortPushes([]interface{}{
		"a1", "b1", "c1",
		"a2", "b2", "c2",
		"a3", "b3", "c3",
		"b4", "c4",
		"b5",
	}, o.Main().Out())
}

func Test_StreamJoin__Uneven_IncompleteStreams(t *testing.T) {
	Init()
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: streamCtrlJoinCfg.blueprint.Id,
			Properties: core.Properties{
				"streams": []interface{}{"a", "b", "c"},
			},
			Generics: core.Generics{
				"itemType": {Type: "string"},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()

	o.Main().In().Map("stream_a").PushBOS()
	o.Main().In().Map("stream_a").Stream().Push("a1")
	o.Main().In().Map("stream_a").Stream().Push("a2")
	o.Main().In().Map("stream_a").Stream().Push("a3")
	o.Main().In().Map("stream_b").PushBOS()
	o.Main().In().Map("stream_b").Stream().Push("b1")
	o.Main().In().Map("stream_b").Stream().Push("b2")
	o.Main().In().Map("stream_b").Stream().Push("b3")
	o.Main().In().Map("stream_b").Stream().Push("b4")
	o.Main().In().Map("stream_b").Stream().Push("b5")
	o.Main().In().Map("stream_c").PushBOS()
	o.Main().In().Map("stream_c").Stream().Push("c1")
	o.Main().In().Map("stream_c").Stream().Push("c2")
	o.Main().In().Map("stream_c").Stream().Push("c3")
	o.Main().In().Map("stream_c").Stream().Push("c4")


	a.True(core.IsBOS(o.Main().Out().Stream().Pull()))
	a.PortPushes("a1", o.Main().Out().Stream())
	a.PortPushes("b1", o.Main().Out().Stream())
	a.PortPushes("c1", o.Main().Out().Stream())
	a.PortPushes("a2", o.Main().Out().Stream())
	a.PortPushes("b2", o.Main().Out().Stream())
	a.PortPushes("c2", o.Main().Out().Stream())
	a.PortPushes("a3", o.Main().Out().Stream())
	a.PortPushes("b3", o.Main().Out().Stream())
	a.PortPushes("c3", o.Main().Out().Stream())
	a.PortPushes("b4", o.Main().Out().Stream())
	a.PortPushes("c4", o.Main().Out().Stream())
	a.PortPushes("b5", o.Main().Out().Stream())
}

func Test_StreamJoin__FirstStreamEmpty_IncompleteStreams(t *testing.T) {
	Init()
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: streamCtrlJoinCfg.blueprint.Id,
			Properties: core.Properties{
				"streams": []interface{}{"a", "b", "c"},
			},
			Generics: core.Generics{
				"itemType": {Type: "string"},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()

	o.Main().In().Map("stream_a").PushBOS()
	o.Main().In().Map("stream_b").PushBOS()
	o.Main().In().Map("stream_b").Stream().Push("b1")
	o.Main().In().Map("stream_b").Stream().Push("b2")
	o.Main().In().Map("stream_b").Stream().Push("b3")
	o.Main().In().Map("stream_b").Stream().Push("b4")
	o.Main().In().Map("stream_b").Stream().Push("b5")
	o.Main().In().Map("stream_c").PushBOS()
	o.Main().In().Map("stream_c").Stream().Push("c1")
	o.Main().In().Map("stream_c").Stream().Push("c2")
	o.Main().In().Map("stream_c").Stream().Push("c3")
	o.Main().In().Map("stream_c").Stream().Push("c4")


	a.True(core.IsBOS(o.Main().Out().Stream().Pull()))
	a.PortPushes("b1", o.Main().Out().Stream())
	a.PortPushes("c1", o.Main().Out().Stream())
	a.PortPushes("b2", o.Main().Out().Stream())
	a.PortPushes("c2", o.Main().Out().Stream())
	a.PortPushes("b3", o.Main().Out().Stream())
	a.PortPushes("c3", o.Main().Out().Stream())
	a.PortPushes("b4", o.Main().Out().Stream())
	a.PortPushes("c4", o.Main().Out().Stream())
	a.PortPushes("b5", o.Main().Out().Stream())
}