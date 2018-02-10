package tests

import (
	"github.com/stretchr/testify/require"
	"slang"
	"slang/core"
	"slang/tests/assertions"
	"testing"
)

func TestOperator_ReadOperator_1_OuterOperator(t *testing.T) {
	a := assertions.New(t)
	o, err := slang.BuildOperator("test_data/voidOp.json", false)
	a.NoError(err)
	a.True(o.In().Connected(o.Out()))

	o.Out().Bufferize()
	o.In().Push("hallo")

	a.PortPushes([]interface{}{"hallo"}, o.Out())
}

func TestOperator_ReadOperator_UnknownOperator(t *testing.T) {
	a := assertions.New(t)
	_, err := slang.BuildOperator(`test_data/unknownOp.json`, false)
	a.Error(err)
}

func TestOperator_ReadOperator_1_BuiltinOperator_Eval(t *testing.T) {
	a := assertions.New(t)
	o, err := slang.BuildOperator("test_data/usingBuiltinOp.json", false)
	a.NoError(err)

	oPasser := o.Child("passer")
	a.NotNil(oPasser)
	a.True(o.In().Connected(oPasser.In()))
	a.True(oPasser.Out().Connected(o.Out()))

	o.Out().Bufferize()
	o.In().Push(map[string]interface{}{"a": "hallo"})

	o.Start()

	a.PortPushes([]interface{}{"hallo"}, o.Out())
}

func TestOperator_ReadOperator_NestedOperator_1_Child(t *testing.T) {
	a := assertions.New(t)
	o, err := slang.BuildOperator("test_data/nested_op/usingCustomOp1.json", false)
	a.NoError(err)

	o.Out().Bufferize()
	o.In().Push("hallo")

	o.Start()

	a.PortPushes([]interface{}{"hallo"}, o.Out())
}

func TestOperator_ReadOperator_NestedOperator_N_Child(t *testing.T) {
	a := assertions.New(t)
	o, err := slang.BuildOperator("test_data/nested_op/usingCustomOpN.json", false)
	a.NoError(err)

	o.Out().Bufferize()
	o.In().Push("hallo")

	o.Start()

	a.PortPushes([]interface{}{"hallo"}, o.Out())
}

func TestOperator_ReadOperator_NestedOperator_SubChild(t *testing.T) {
	a := assertions.New(t)
	o, err := slang.BuildOperator("test_data/nested_op/usingSubCustomOpDouble.json", false)
	a.NoError(err)

	o.Out().Bufferize()
	o.In().Push("hallo")
	o.In().Push(2.0)

	o.Start()

	a.PortPushes([]interface{}{"hallohallo", 4.0}, o.Out())
}

func TestOperator_ReadOperator_NestedOperator_Cwd(t *testing.T) {
	a := assertions.New(t)
	o, err := slang.BuildOperator("test_data/cwdOp.json", false)
	a.NoError(err)

	o.Out().Bufferize()
	o.In().Push("hey")
	o.In().Push(false)

	o.Start()

	a.PortPushes([]interface{}{"hey", false}, o.Out())
}

func TestOperator_ReadOperator__Recursion(t *testing.T) {
	a := assertions.New(t)
	o, err := slang.BuildOperator("test_data/recOp1.json", false)
	a.Error(err)
	a.Nil(o)
}

func TestOperator_ReadOperator_NestedGeneric(t *testing.T) {
	a := assertions.New(t)
	o, err := slang.BuildOperator("test_data/nested_generic/main.json", false)
	require.NoError(t, err)

	o.Out().Bufferize()
	o.In().Push("hallo")

	a.PortPushes([]interface{}{"hallo"}, o.Out().Map("left"))
	a.PortPushes([]interface{}{"hallo"}, o.Out().Map("right"))
}

func TestParsePortReference__NilOperator(t *testing.T) {
	a := assertions.New(t)
	p, err := slang.ParsePortReference("test.in", nil)
	a.Error(err)
	a.Nil(p)
}

func TestParsePortReference__NilConnection(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"})
	o2, _ := core.NewOperator("o2", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"})
	o2.SetParent(o1)
	p, err := slang.ParsePortReference("", o1)
	a.Error(err)
	a.Nil(p)
}

func TestParsePortReference__SelfIn(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"})
	p, err := slang.ParsePortReference("(", o1)
	a.NoError(err)

	if p != o1.In() {
		t.Error("wrong port")
	}
}

func TestParsePortReference__SelfOut(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"})
	p, err := slang.ParsePortReference(")", o1)
	a.NoError(err)

	if p != o1.Out() {
		t.Error("wrong port")
	}
}

func TestParsePortReference__SingleIn(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"})
	o2, _ := core.NewOperator("o2", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"})
	o2.SetParent(o1)
	p, err := slang.ParsePortReference("(o2", o1)
	a.NoError(err)

	if p != o2.In() {
		t.Error("wrong port")
	}
}

func TestParsePortReference__SingleOut(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"})
	o2, _ := core.NewOperator("o2", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"})
	o2.SetParent(o1)
	p, err := slang.ParsePortReference("o2)", o1)
	a.NoError(err)

	if p != o2.Out() {
		t.Error("wrong port")
	}
}

func TestParsePortReference__Map(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"})
	o2, _ := core.NewOperator("o2", nil, core.PortDef{Type: "map", Map: map[string]*core.PortDef{"a": {Type: "number"}}}, core.PortDef{Type: "number"})
	o2.SetParent(o1)
	p, err := slang.ParsePortReference("a(o2", o1)
	a.NoError(err)

	if p != o2.In().Map("a") {
		t.Error("wrong port")
	}
}

func TestParsePortReference__Map__UnknownKey(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"})
	o2, _ := core.NewOperator("o2", nil, core.PortDef{Type: "map", Map: map[string]*core.PortDef{"a": {Type: "number"}}}, core.PortDef{Type: "number"})
	o2.SetParent(o1)
	p, err := slang.ParsePortReference("b(o2", o1)
	a.Error(err)
	a.Nil(p)
}

func TestParsePortReference__Map__DescendingTooDeep(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"})
	o2, _ := core.NewOperator("o2", nil, core.PortDef{Type: "map", Map: map[string]*core.PortDef{"a": {Type: "number"}}}, core.PortDef{Type: "number"})
	o2.SetParent(o1)
	p, err := slang.ParsePortReference("b.c(o2", o1)
	a.Error(err)
	a.Nil(p)
}

func TestParsePortReference__NestedMap(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"})
	o2, _ := core.NewOperator("o2", nil, core.PortDef{Type: "map", Map: map[string]*core.PortDef{"a": {Type: "map", Map: map[string]*core.PortDef{"b": {Type: "number"}}}}}, core.PortDef{Type: "number"})
	o2.SetParent(o1)
	p, err := slang.ParsePortReference("a.b(o2", o1)
	a.NoError(err)

	if p != o2.In().Map("a").Map("b") {
		t.Error("wrong port")
	}
}

func TestParsePortReference__Stream(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"})
	o2, _ := core.NewOperator("o2", nil, core.PortDef{Type: "stream", Stream: &core.PortDef{Type: "number"}}, core.PortDef{Type: "number"})
	o2.SetParent(o1)
	p, err := slang.ParsePortReference(".(o2", o1)
	a.NoError(err)

	if p != o2.In().Stream() {
		t.Error("wrong port")
	}
}

func TestParsePortReference__StreamMap(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"})
	o2, _ := core.NewOperator("o2", nil,
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
		core.PortDef{Type: "number"})
	o2.SetParent(o1)
	p, err := slang.ParsePortReference("..a..a.(o2", o1)
	a.NoError(err)
	a.Equal(p, o2.In().Stream().Map("a").Stream().Map("a").Stream(), "wrong port")
}
