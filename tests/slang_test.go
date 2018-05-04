package tests

import (
	"github.com/stretchr/testify/require"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"testing"
	"github.com/Bitspark/slang/pkg/api"
)

var e = api.NewEnviron("./")

func TestOperator_ReadOperator_1_OuterOperator(t *testing.T) {
	a := assertions.New(t)
	o, err := e.BuildOperator("test_data/voidOp.json", nil, nil, false)
	a.NoError(err)
	a.True(o.In().Connected(o.Out()))

	o.Out().Bufferize()
	o.In().Push("hallo")

	a.PortPushesAll([]interface{}{"hallo"}, o.Out())
}

func TestOperator_ReadOperator_UnknownOperator(t *testing.T) {
	a := assertions.New(t)
	_, err := e.BuildOperator(`test_data/unknownOp.json`, nil, nil, false)
	a.Error(err)
}

func TestOperator_ReadOperator_1_BuiltinOperator_Eval(t *testing.T) {
	a := assertions.New(t)
	o, err := e.BuildOperator("test_data/usingBuiltinOp.json", nil, nil, false)
	a.NoError(err)

	oPasser := o.Child("passer")
	a.NotNil(oPasser)
	a.True(o.In().Connected(oPasser.In()))
	a.True(oPasser.Out().Connected(o.Out()))

	o.Out().Bufferize()
	o.In().Push(map[string]interface{}{"a": "hallo"})

	o.Start()

	a.PortPushesAll([]interface{}{"hallo"}, o.Out())
}

func TestOperator_ReadOperator_NestedOperator_1_Child(t *testing.T) {
	a := assertions.New(t)
	o, err := e.BuildOperator("test_data/nested_op/usingCustomOp1.json", nil, nil, false)
	a.NoError(err)

	o.Out().Bufferize()
	o.In().Push("hallo")

	o.Start()

	a.PortPushesAll([]interface{}{"hallo"}, o.Out())
}

func TestOperator_ReadOperator_NestedOperator_N_Child(t *testing.T) {
	a := assertions.New(t)
	o, err := e.BuildOperator("test_data/nested_op/usingCustomOpN.json", nil, nil, false)
	a.NoError(err)

	o.Out().Bufferize()
	o.In().Push("hallo")

	o.Start()

	a.PortPushesAll([]interface{}{"hallo"}, o.Out())
}

func TestOperator_ReadOperator_NestedOperator_SubChild(t *testing.T) {
	a := assertions.New(t)
	o, err := e.BuildOperator("test_data/nested_op/usingSubCustomOpDouble.json", nil, nil, false)
	a.NoError(err)

	o.Out().Bufferize()
	o.In().Push("hallo")
	o.In().Push(2.0)

	o.Start()

	a.PortPushesAll([]interface{}{"hallohallo", 4.0}, o.Out())
}

func TestOperator_ReadOperator_NestedOperator_Cwd(t *testing.T) {
	a := assertions.New(t)
	o, err := e.BuildOperator("test_data/cwdOp.json", nil, nil, false)
	a.NoError(err)

	o.Out().Bufferize()
	o.In().Push("hey")
	o.In().Push(false)

	o.Start()

	a.PortPushesAll([]interface{}{"hey", false}, o.Out())
}

func TestOperator_ReadOperator__Recursion(t *testing.T) {
	a := assertions.New(t)
	o, err := e.BuildOperator("test_data/recOp1.json", nil, nil, false)
	a.Error(err)
	a.Nil(o)
}

func TestOperator_ReadOperator_NestedGeneric(t *testing.T) {
	a := assertions.New(t)
	o, err := e.BuildOperator("test_data/nested_generic/main.json", nil, nil, false)
	require.NoError(t, err)

	o.Out().Bufferize()
	o.In().Push("hallo")

	a.PortPushesAll([]interface{}{"hallo"}, o.Out().Map("left"))
	a.PortPushesAll([]interface{}{"hallo"}, o.Out().Map("right"))
}

func TestParsePortReference__NilOperator(t *testing.T) {
	a := assertions.New(t)
	p, err := api.ParsePortReference("test.in", nil)
	a.Error(err)
	a.Nil(p)
}

func TestParsePortReference__NilConnection(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	o2, _ := core.NewOperator("o2", nil, nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	o2.SetParent(o1)
	p, err := api.ParsePortReference("", o1)
	a.Error(err)
	a.Nil(p)
}

func TestParsePortReference__SelfIn(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	p, err := api.ParsePortReference("(", o1)
	a.NoError(err)
	a.Equal(o1.In(), p, "wrong port")
}

func TestParsePortReference__SelfOut(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	p, err := api.ParsePortReference(")", o1)
	a.NoError(err)
	a.Equal(o1.Out(), p, "wrong port")
}

func TestParsePortReference__SingleIn(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	o2, _ := core.NewOperator("o2", nil, nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	o2.SetParent(o1)
	p, err := api.ParsePortReference("(o2", o1)
	a.NoError(err)
	a.Equal(o2.In(), p, "wrong port")
}

func TestParsePortReference__SingleOut(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	o2, _ := core.NewOperator("o2", nil, nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	o2.SetParent(o1)
	p, err := api.ParsePortReference("o2)", o1)
	a.NoError(err)
	a.Equal(o2.Out(), p, "wrong port")
}

func TestParsePortReference__Map(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	o2, _ := core.NewOperator("o2", nil, nil, core.PortDef{Type: "map", Map: map[string]*core.PortDef{"a": {Type: "number"}}}, core.PortDef{Type: "number"}, nil)
	o2.SetParent(o1)
	p, err := api.ParsePortReference("a(o2", o1)
	a.NoError(err)
	a.Equal(o2.In().Map("a"), p, "wrong port")
}

func TestParsePortReference__Map__UnknownKey(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	o2, _ := core.NewOperator("o2", nil, nil, core.PortDef{Type: "map", Map: map[string]*core.PortDef{"a": {Type: "number"}}}, core.PortDef{Type: "number"}, nil)
	o2.SetParent(o1)
	p, err := api.ParsePortReference("b(o2", o1)
	a.Error(err)
	a.Nil(p)
}

func TestParsePortReference__Map__DescendingTooDeep(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	o2, _ := core.NewOperator("o2", nil, nil, core.PortDef{Type: "map", Map: map[string]*core.PortDef{"a": {Type: "number"}}}, core.PortDef{Type: "number"}, nil)
	o2.SetParent(o1)
	p, err := api.ParsePortReference("b.c(o2", o1)
	a.Error(err)
	a.Nil(p)
}

func TestParsePortReference__NestedMap(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	o2, _ := core.NewOperator("o2", nil, nil, core.PortDef{Type: "map", Map: map[string]*core.PortDef{"a": {Type: "map", Map: map[string]*core.PortDef{"b": {Type: "number"}}}}}, core.PortDef{Type: "number"}, nil)
	o2.SetParent(o1)
	p, err := api.ParsePortReference("a.b(o2", o1)
	a.NoError(err)
	a.Equal(o2.In().Map("a").Map("b"), p, "wrong port")
}

func TestParsePortReference__Stream(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	o2, _ := core.NewOperator("o2", nil, nil, core.PortDef{Type: "stream", Stream: &core.PortDef{Type: "number"}}, core.PortDef{Type: "number"}, nil)
	o2.SetParent(o1)
	p, err := api.ParsePortReference("~(o2", o1)
	a.NoError(err)
	a.Equal(o2.In().Stream(), p, "wrong port")
}

func TestParsePortReference__StreamMap(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	o2, _ := core.NewOperator("o2", nil, nil,
		core.PortDef{
			Type: "stream",
			Stream: &core.PortDef{
				Type: "map",
				Map: map[string]*core.PortDef{
					"a": {
						Type: "stream",
						Stream: &core.PortDef{
							Type: "map",
							Map: map[string]*core.PortDef{
								"a": {
									Type: "stream",
									Stream: &core.PortDef{
										Type: "boolean",
									},
								},
							},
						},
					},
				},
			},
		},
		core.PortDef{Type: "number"},
		nil)
	o2.SetParent(o1)
	p, err := api.ParsePortReference("~.a.~.a.~(o2", o1)
	a.NoError(err)
	a.Equal(o2.In().Stream().Map("a").Stream().Map("a").Stream(), p, "wrong port")
}

func TestParsePortReference__Delegates_In(t *testing.T) {
	a := assertions.New(t)
	o, _ := core.NewOperator(
		"o1",
		nil, nil,
		core.PortDef{Type: "number"},
		core.PortDef{Type: "number"},
		map[string]*core.DelegateDef{
			"test": {
				In:  core.PortDef{Type: "number"},
				Out: core.PortDef{Type: "number"}},
		})
	p, err := api.ParsePortReference("(.test", o)
	a.NoError(err)
	a.Equal(o.Delegate("test").In(), p, "wrong port")
}

func TestParsePortReference__Delegates_Out(t *testing.T) {
	a := assertions.New(t)
	o, _ := core.NewOperator(
		"o1",
		nil, nil,
		core.PortDef{Type: "number"},
		core.PortDef{Type: "number"},
		map[string]*core.DelegateDef{
			"test": {
				In:  core.PortDef{Type: "number"},
				Out: core.PortDef{Type: "number"}},
		})
	p, err := api.ParsePortReference(".test)", o)
	a.NoError(err)
	a.Equal(o.Delegate("test").Out(), p, "wrong port")
}

func TestParsePortReference__Delegates_SingleIn(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator(
		"o1",
		nil, nil,
		core.PortDef{Type: "number"},
		core.PortDef{Type: "number"},
		nil)
	o2, _ := core.NewOperator(
		"o2",
		nil, nil,
		core.PortDef{Type: "number"},
		core.PortDef{Type: "number"},
		map[string]*core.DelegateDef{
			"test": {
				In:  core.PortDef{Type: "number"},
				Out: core.PortDef{Type: "number"}},
		})
	o2.SetParent(o1)
	p, err := api.ParsePortReference("(o2.test", o1)
	a.NoError(err)
	a.Equal(o2.Delegate("test").In(), p, "wrong port")
}

func TestParsePortReference__Delegates_SingleOut(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator(
		"o1",
		nil, nil,
		core.PortDef{Type: "number"},
		core.PortDef{Type: "number"},
		nil)
	o2, _ := core.NewOperator(
		"o2",
		nil, nil,
		core.PortDef{Type: "number"},
		core.PortDef{Type: "number"},
		map[string]*core.DelegateDef{
			"test": {
				In:  core.PortDef{Type: "number"},
				Out: core.PortDef{Type: "number"}},
		})
	o2.SetParent(o1)
	p, err := api.ParsePortReference("o2.test)", o1)
	a.NoError(err)
	a.Equal(o2.Delegate("test").Out(), p, "wrong port")
}

func TestParsePortReference__Delegates_Map(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator(
		"o1", nil, nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	o2, _ := core.NewOperator(
		"o2",
		nil, nil,
		core.PortDef{Type: "number"}, core.PortDef{Type: "number"},
		map[string]*core.DelegateDef{
			"test": {
				In:  core.PortDef{Type: "map", Map: map[string]*core.PortDef{"a": {Type: "number"}}},
				Out: core.PortDef{Type: "number"}},
		})
	o2.SetParent(o1)
	p, err := api.ParsePortReference("a(o2.test", o1)
	a.NoError(err)
	a.Equal(o2.Delegate("test").In().Map("a"), p, "wrong port")
}
