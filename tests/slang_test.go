package tests

import (
	"slang"
	"slang/core"
	"slang/tests/assertions"
	"testing"
)

func TestOperator_ReadOperator_1_OuterOperator(t *testing.T) {
	a := assertions.New(t)
	o, err := slang.ReadOperator("test_data/voidOp.json")
	a.NoError(err)
	a.True(o.In().Connected(o.Out()))

	o.Out().Bufferize()
	o.In().Push("hallo")

	a.PortPushes([]interface{}{"hallo"}, o.Out())
}

func TestOperator_ReadOperator_UnknownOperator(t *testing.T) {
	a := assertions.New(t)
	_, err := slang.ReadOperator(`test_data/unknownOp.json`)
	a.Error(err)
}

func TestOperator_ReadOperator_1_BuiltinOperator_Eval(t *testing.T) {
	a := assertions.New(t)
	o, err := slang.ReadOperator("test_data/usingBuiltinOp.json")
	a.NoError(err)

	oPasser := o.Child("passer")
	a.NotNil(oPasser)
	a.True(o.In().Connected(oPasser.In().Map("a")))
	a.True(oPasser.Out().Connected(o.Out()))

	o.Out().Bufferize()
	o.In().Push("hallo")

	o.Start()

	a.PortPushes([]interface{}{"hallo"}, o.Out())
}

func TestOperator_ReadOperator_NestedOperator_1_Child(t *testing.T) {
	a := assertions.New(t)
	o, err := slang.ReadOperator("test_data/nested_op/usingCustomOp1.json")
	a.NoError(err)

	o.Out().Bufferize()
	o.In().Push("hallo")

	o.Start()

	a.PortPushes([]interface{}{"hallo"}, o.Out())
}

func TestOperator_ReadOperator_NestedOperator_N_Child(t *testing.T) {
	a := assertions.New(t)
	o, err := slang.ReadOperator("test_data/nested_op/usingCustomOpN.json")
	a.NoError(err)

	o.Out().Bufferize()
	o.In().Push("hallo")

	o.Start()

	a.PortPushes([]interface{}{"hallo"}, o.Out())
}

func TestOperator_ReadOperator_NestedOperator_SubChild(t *testing.T) {
	a := assertions.New(t)
	o, err := slang.ReadOperator("test_data/nested_op/usingSubCustomOpDouble.json")
	a.NoError(err)

	o.Out().Bufferize()
	o.In().Push("hallo")
	o.In().Push(2.0)

	o.Start()

	a.PortPushes([]interface{}{"hallohallo", 4.0}, o.Out())
}

func TestOperator_ReadOperator_NestedOperator_Cwd(t *testing.T) {
	a := assertions.New(t)
	o, err := slang.ReadOperator("test_data/cwdOp.json")
	a.NoError(err)

	o.Out().Bufferize()
	o.In().Push("hey")
	o.In().Push(false)

	o.Start()

	a.PortPushes([]interface{}{"hey", false}, o.Out())
}

func TestOperator_ReadOperator__Recursion(t *testing.T) {
	a := assertions.New(t)
	o, err := slang.ReadOperator("test_data/recOp1.json")
	a.Error(err)
	a.Nil(o)
}

func TestParseConnection__NilOperator(t *testing.T) {
	a := assertions.New(t)
	p, err := slang.ParsePort("test.in", nil)
	a.Error(err)
	a.Nil(p)
}

func TestParseConnection__NilConnection(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	core.NewOperator("o2", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, o1)
	p, err := slang.ParsePort("", o1)
	a.Error(err)
	a.Nil(p)
}

func TestParseConnection__SelfIn(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	p, err := slang.ParsePort(":in", o1)
	a.NoError(err)

	if p != o1.In() {
		t.Error("wrong port")
	}
}

func TestParseConnection__SelfOut(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	p, err := slang.ParsePort(":out", o1)
	a.NoError(err)

	if p != o1.Out() {
		t.Error("wrong port")
	}
}

func TestParseConnection__SingleIn(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	o2, _ := core.NewOperator("o2", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, o1)
	p, err := slang.ParsePort("o2:in", o1)
	a.NoError(err)

	if p != o2.In() {
		t.Error("wrong port")
	}
}

func TestParseConnection__SingleOut(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	o2, _ := core.NewOperator("o2", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, o1)
	p, err := slang.ParsePort("o2:out", o1)
	a.NoError(err)

	if p != o2.Out() {
		t.Error("wrong port")
	}
}

func TestParseConnection__Map(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	o2, _ := core.NewOperator("o2", nil, core.PortDef{Type: "map", Map: map[string]core.PortDef{"a": {Type: "number"}}}, core.PortDef{Type: "number"}, o1)
	p, err := slang.ParsePort("o2:in.a", o1)
	a.NoError(err)

	if p != o2.In().Map("a") {
		t.Error("wrong port")
	}
}

func TestParseConnection__Map__UnknownKey(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	core.NewOperator("o2", nil, core.PortDef{Type: "map", Map: map[string]core.PortDef{"a": {Type: "number"}}}, core.PortDef{Type: "number"}, o1)
	p, err := slang.ParsePort("o2:in.b", o1)
	a.Error(err)
	a.Nil(p)
}

func TestParseConnection__Map__DescendingTooDeep(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	core.NewOperator("o2", nil, core.PortDef{Type: "map", Map: map[string]core.PortDef{"a": {Type: "number"}}}, core.PortDef{Type: "number"}, o1)
	p, err := slang.ParsePort("o2:in.b.c", o1)
	a.Error(err)
	a.Nil(p)
}

func TestParseConnection__NestedMap(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	o2, _ := core.NewOperator("o2", nil, core.PortDef{Type: "map", Map: map[string]core.PortDef{"a": {Type: "map", Map: map[string]core.PortDef{"b": {Type: "number"}}}}}, core.PortDef{Type: "number"}, o1)
	p, err := slang.ParsePort("o2:in.a.b", o1)
	a.NoError(err)

	if p != o2.In().Map("a").Map("b") {
		t.Error("wrong port")
	}
}

func TestParseConnection__Stream(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	o2, _ := core.NewOperator("o2", nil, core.PortDef{Type: "stream", Stream: &core.PortDef{Type: "number"}}, core.PortDef{Type: "number"}, o1)
	p, err := slang.ParsePort("o2:in", o1)
	a.NoError(err)

	if p != o2.In().Stream() {
		t.Error("wrong port")
	}
}

func TestParseConnection__StreamMap(t *testing.T) {
	a := assertions.New(t)
	o1, _ := core.NewOperator("o1", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	o2, _ := core.NewOperator("o2", nil,
		core.PortDef{
			Type: "stream",
			Stream: &core.PortDef{
				Type: "map",
				Map: map[string]core.PortDef{
					"a": {
						Type: "stream",
						Stream: &core.PortDef{
							Type: "map",
							Map: map[string]core.PortDef{
								"a": {
									Type: "stream",
									Stream: &core.PortDef{
										Type: "boolean",
									},
								},
							},
						},
					},
				}},
		},
		core.PortDef{Type: "number"}, o1)
	p, err := slang.ParsePort("o2:in.a.a", o1)
	a.NoError(err)

	if p != o2.In().Stream().Map("a").Stream().Map("a").Stream() {
		t.Error("wrong port")
	}
}
