package tests

import (
	"slang"
	"slang/assertions"
	"slang/op"
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

func TestOperator_ReadOperator_1_BuiltinOperator_Function(t *testing.T) {
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
	p, err := slang.ParseConnection("test.in", nil)
	a.Error(err)
	a.Nil(p)
}

func TestParseConnection__NilConnection(t *testing.T) {
	a := assertions.New(t)
	o1, _ := op.NewOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	op.NewOperator("o2", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, o1)
	p, err := slang.ParseConnection("", o1)
	a.Error(err)
	a.Nil(p)
}

func TestParseConnection__SelfIn(t *testing.T) {
	a := assertions.New(t)
	o1, _ := op.NewOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	p, err := slang.ParseConnection(":in", o1)
	a.NoError(err)

	if p != o1.In() {
		t.Error("wrong port")
	}
}

func TestParseConnection__SelfOut(t *testing.T) {
	a := assertions.New(t)
	o1, _ := op.NewOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	p, err := slang.ParseConnection(":out", o1)
	a.NoError(err)

	if p != o1.Out() {
		t.Error("wrong port")
	}
}

func TestParseConnection__SingleIn(t *testing.T) {
	a := assertions.New(t)
	o1, _ := op.NewOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	o2, _ := op.NewOperator("o2", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, o1)
	p, err := slang.ParseConnection("o2:in", o1)
	a.NoError(err)

	if p != o2.In() {
		t.Error("wrong port")
	}
}

func TestParseConnection__SingleOut(t *testing.T) {
	a := assertions.New(t)
	o1, _ := op.NewOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	o2, _ := op.NewOperator("o2", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, o1)
	p, err := slang.ParseConnection("o2:out", o1)
	a.NoError(err)

	if p != o2.Out() {
		t.Error("wrong port")
	}
}

func TestParseConnection__Map(t *testing.T) {
	a := assertions.New(t)
	o1, _ := op.NewOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	o2, _ := op.NewOperator("o2", nil, op.PortDef{Type: "map", Map: map[string]op.PortDef{"a": {Type: "number"}}}, op.PortDef{Type: "number"}, o1)
	p, err := slang.ParseConnection("o2:in.a", o1)
	a.NoError(err)

	if p != o2.In().Map("a") {
		t.Error("wrong port")
	}
}

func TestParseConnection__Map__UnknownKey(t *testing.T) {
	a := assertions.New(t)
	o1, _ := op.NewOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	op.NewOperator("o2", nil, op.PortDef{Type: "map", Map: map[string]op.PortDef{"a": {Type: "number"}}}, op.PortDef{Type: "number"}, o1)
	p, err := slang.ParseConnection("o2:in.b", o1)
	a.Error(err)
	a.Nil(p)
}

func TestParseConnection__Map__DescendingTooDeep(t *testing.T) {
	a := assertions.New(t)
	o1, _ := op.NewOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	op.NewOperator("o2", nil, op.PortDef{Type: "map", Map: map[string]op.PortDef{"a": {Type: "number"}}}, op.PortDef{Type: "number"}, o1)
	p, err := slang.ParseConnection("o2:in.b.c", o1)
	a.Error(err)
	a.Nil(p)
}

func TestParseConnection__NestedMap(t *testing.T) {
	a := assertions.New(t)
	o1, _ := op.NewOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	o2, _ := op.NewOperator("o2", nil, op.PortDef{Type: "map", Map: map[string]op.PortDef{"a": {Type: "map", Map: map[string]op.PortDef{"b": {Type: "number"}}}}}, op.PortDef{Type: "number"}, o1)
	p, err := slang.ParseConnection("o2:in.a.b", o1)
	a.NoError(err)

	if p != o2.In().Map("a").Map("b") {
		t.Error("wrong port")
	}
}

func TestParseConnection__Stream(t *testing.T) {
	a := assertions.New(t)
	o1, _ := op.NewOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	o2, _ := op.NewOperator("o2", nil, op.PortDef{Type: "stream", Stream: &op.PortDef{Type: "number"}}, op.PortDef{Type: "number"}, o1)
	p, err := slang.ParseConnection("o2:in", o1)
	a.NoError(err)

	if p != o2.In().Stream() {
		t.Error("wrong port")
	}
}

func TestParseConnection__StreamMap(t *testing.T) {
	a := assertions.New(t)
	o1, _ := op.NewOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	o2, _ := op.NewOperator("o2", nil,
		op.PortDef{
			Type: "stream",
			Stream: &op.PortDef{
				Type: "map",
				Map: map[string]op.PortDef{
					"a": {
						Type: "stream",
						Stream: &op.PortDef{
							Type: "map",
							Map: map[string]op.PortDef{
								"a": {
									Type: "stream",
									Stream: &op.PortDef{
										Type: "boolean",
									},
								},
							},
						},
					},
				}},
		},
		op.PortDef{Type: "number"}, o1)
	p, err := slang.ParseConnection("o2:in.a.a", o1)
	a.NoError(err)

	if p != o2.In().Stream().Map("a").Stream().Map("a").Stream() {
		t.Error("wrong port")
	}
}
