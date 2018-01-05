package tests

import (
	"slang"
	"slang/op"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOperator_ReadOperator_1_OuterOperator(t *testing.T) {
	o, err := slang.ReadOperator("test_data/voidOp.json")
	assert.NoError(t, err)
	assert.True(t, o.In().Connected(o.Out()))

	o.Out().Bufferize()
	o.In().Push("hallo")

	AssertPortItems(t, []interface{}{"hallo"}, o.Out())
}

func TestOperator_ReadOperator_UnknownOperator(t *testing.T) {
	_, err := slang.ReadOperator(`test_data/unknownOp.json`)
	assert.Error(t, err)
}

func TestOperator_ReadOperator_1_BuiltinOperator_Function(t *testing.T) {
	o, err := slang.ReadOperator("test_data/usingBuiltinOp.json")
	assert.NoError(t, err)

	oPasser := o.Child("passer")
	assert.NotNil(t, oPasser)
	assert.True(t, o.In().Connected(oPasser.In().Map("a")))
	assert.True(t, oPasser.Out().Connected(o.Out()))

	o.Out().Bufferize()
	o.In().Push("hallo")

	o.Start()

	AssertPortItems(t, []interface{}{"hallo"}, o.Out())
}

func TestOperator_ReadOperator_NestedOperator_1_Child(t *testing.T) {
	o, err := slang.ReadOperator("test_data/nested_op/usingCustomOp1.json")
	assert.NoError(t, err)

	o.Out().Bufferize()
	o.In().Push("hallo")

	o.Start()

	AssertPortItems(t, []interface{}{"hallo"}, o.Out())
}

func TestOperator_ReadOperator_NestedOperator_N_Child(t *testing.T) {
	o, err := slang.ReadOperator("test_data/nested_op/usingCustomOpN.json")
	assert.NoError(t, err)

	o.Out().Bufferize()
	o.In().Push("hallo")

	o.Start()

	AssertPortItems(t, []interface{}{"hallo"}, o.Out())
}

func TestOperator_ReadOperator_NestedOperator_SubChild(t *testing.T) {
	o, err := slang.ReadOperator("test_data/nested_op/usingSubCustomOpDouble.json")
	assert.NoError(t, err)

	o.Out().Bufferize()
	o.In().Push("hallo")
	o.In().Push(2.0)

	o.Start()

	AssertPortItems(t, []interface{}{"hallohallo", 4.0}, o.Out())
}

func TestOperator_ReadOperator_NestedOperator_Cwd(t *testing.T) {
	o, err := slang.ReadOperator("test_data/cwdOp.json")
	assert.NoError(t, err)

	o.Out().Bufferize()
	o.In().Push("hey")
	o.In().Push(false)

	o.Start()

	AssertPortItems(t, []interface{}{"hey", false}, o.Out())
}

func TestOperator_ReadOperator__Recursion(t *testing.T) {
	a := assert.New(t)
	o, err := slang.ReadOperator("test_data/recOp1.json")
	a.Error(err)
	a.Nil(o)
}

func TestParseConnection__NilOperator(t *testing.T) {
	p, err := slang.ParseConnection("test.in", nil)
	assert.Error(t, err)
	assert.Nil(t, p)
}

func TestParseConnection__NilConnection(t *testing.T) {
	o1, _ := op.MakeOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	op.MakeOperator("o2", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, o1)
	p, err := slang.ParseConnection("", o1)
	assert.Error(t, err)
	assert.Nil(t, p)
}

func TestParseConnection__SelfIn(t *testing.T) {
	o1, _ := op.MakeOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	p, err := slang.ParseConnection(":in", o1)
	assert.NoError(t, err)

	if p != o1.In() {
		t.Error("wrong port")
	}
}

func TestParseConnection__SelfOut(t *testing.T) {
	o1, _ := op.MakeOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	p, err := slang.ParseConnection(":out", o1)
	assert.NoError(t, err)

	if p != o1.Out() {
		t.Error("wrong port")
	}
}

func TestParseConnection__SingleIn(t *testing.T) {
	o1, _ := op.MakeOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	o2, _ := op.MakeOperator("o2", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, o1)
	p, err := slang.ParseConnection("o2:in", o1)
	assert.NoError(t, err)

	if p != o2.In() {
		t.Error("wrong port")
	}
}

func TestParseConnection__SingleOut(t *testing.T) {
	o1, _ := op.MakeOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	o2, _ := op.MakeOperator("o2", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, o1)
	p, err := slang.ParseConnection("o2:out", o1)
	assert.NoError(t, err)

	if p != o2.Out() {
		t.Error("wrong port")
	}
}

func TestParseConnection__Map(t *testing.T) {
	o1, _ := op.MakeOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	o2, _ := op.MakeOperator("o2", nil, op.PortDef{Type: "map", Map: map[string]op.PortDef{"a": {Type: "number"}}}, op.PortDef{Type: "number"}, o1)
	p, err := slang.ParseConnection("o2:in.a", o1)
	assert.NoError(t, err)

	if p != o2.In().Map("a") {
		t.Error("wrong port")
	}
}

func TestParseConnection__Map__UnknownKey(t *testing.T) {
	o1, _ := op.MakeOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	op.MakeOperator("o2", nil, op.PortDef{Type: "map", Map: map[string]op.PortDef{"a": {Type: "number"}}}, op.PortDef{Type: "number"}, o1)
	p, err := slang.ParseConnection("o2:in.b", o1)
	assert.Error(t, err)
	assert.Nil(t, p)
}

func TestParseConnection__Map__DescendingTooDeep(t *testing.T) {
	o1, _ := op.MakeOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	op.MakeOperator("o2", nil, op.PortDef{Type: "map", Map: map[string]op.PortDef{"a": {Type: "number"}}}, op.PortDef{Type: "number"}, o1)
	p, err := slang.ParseConnection("o2:in.b.c", o1)
	assert.Error(t, err)
	assert.Nil(t, p)
}

func TestParseConnection__NestedMap(t *testing.T) {
	o1, _ := op.MakeOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	o2, _ := op.MakeOperator("o2", nil, op.PortDef{Type: "map", Map: map[string]op.PortDef{"a": {Type: "map", Map: map[string]op.PortDef{"b": {Type: "number"}}}}}, op.PortDef{Type: "number"}, o1)
	p, err := slang.ParseConnection("o2:in.a.b", o1)
	assert.NoError(t, err)

	if p != o2.In().Map("a").Map("b") {
		t.Error("wrong port")
	}
}

func TestParseConnection__Stream(t *testing.T) {
	o1, _ := op.MakeOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	o2, _ := op.MakeOperator("o2", nil, op.PortDef{Type: "stream", Stream: &op.PortDef{Type: "number"}}, op.PortDef{Type: "number"}, o1)
	p, err := slang.ParseConnection("o2:in", o1)
	assert.NoError(t, err)

	if p != o2.In().Stream() {
		t.Error("wrong port")
	}
}

func TestParseConnection__StreamMap(t *testing.T) {
	o1, _ := op.MakeOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	o2, _ := op.MakeOperator("o2", nil,
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
	assert.NoError(t, err)

	if p != o2.In().Stream().Map("a").Stream().Map("a").Stream() {
		t.Error("wrong port")
	}
}
