package slang

import (
	"slang/op"
	"testing"
)

func TestOperator_ReadOperator_1_OuterOperator(t *testing.T) {
	o, err := ReadOperator("test_data/voidOp.json")
	assertNoError(t, err)
	assertTrue(t, o.InPort().Connected(o.OutPort()))

	o.OutPort().Bufferize()
	o.InPort().Push("hallo")

	assertPortItems(t, []interface{}{"hallo"}, o.OutPort())
}

func TestOperator_ReadOperator_UnknownOperator(t *testing.T) {
	_, err := ReadOperator(`test_data/unknownOp.json`)
	assertError(t, err)
}

func TestOperator_ReadOperator_1_BuiltinOperator_Function(t *testing.T) {
	o, err := ReadOperator("test_data/usingBuiltinOp.json")
	assertNoError(t, err)

	oPasser := o.Child("passer")
	assertNotNil(t, oPasser)
	assertTrue(t, o.InPort().Connected(oPasser.InPort().Port("a")))
	assertTrue(t, oPasser.OutPort().Connected(o.OutPort()))

	o.OutPort().Bufferize()
	o.InPort().Push("hallo")

	o.Start()

	assertPortItems(t, []interface{}{"hallo"}, o.OutPort())
}

func TestOperator_ReadOperator_NestedOperator_1_Child(t *testing.T) {
	o, err := ReadOperator("test_data/nested_op/usingCustomOp1.json")
	assertNoError(t, err)

	o.OutPort().Bufferize()
	o.InPort().Push("hallo")

	o.Start()

	assertPortItems(t, []interface{}{"hallo"}, o.OutPort())
}

func TestOperator_ReadOperator_NestedOperator_N_Child(t *testing.T) {
	o, err := ReadOperator("test_data/nested_op/usingCustomOpN.json")
	assertNoError(t, err)

	o.OutPort().Bufferize()
	o.InPort().Push("hallo")

	o.Start()

	assertPortItems(t, []interface{}{"hallo"}, o.OutPort())
}

func TestOperator_ReadOperator_NestedOperator_SubChild(t *testing.T) {
	o, err := ReadOperator("test_data/nested_op/usingSubCustomOpDouble.json")
	assertNoError(t, err)

	o.OutPort().Bufferize()
	o.InPort().Push("hallo")
	o.InPort().Push(2.0)

	o.Start()

	assertPortItems(t, []interface{}{"hallohallo", 4.0}, o.OutPort())
}

func TestParseConnection__NilOperator(t *testing.T) {
	p, err := parseConnection("test.in", nil)
	assertError(t, err)
	assertNil(t, p)
}

func TestParseConnection__NilConnection(t *testing.T) {
	o1, _ := op.MakeOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	op.MakeOperator("o2", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, o1)
	p, err := parseConnection("", o1)
	assertError(t, err)
	assertNil(t, p)
}

func TestParseConnection__SelfIn(t *testing.T) {
	o1, _ := op.MakeOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	p, err := parseConnection(":in", o1)
	assertNoError(t, err)

	if p != o1.InPort() {
		t.Error("wrong port")
	}
}

func TestParseConnection__SelfOut(t *testing.T) {
	o1, _ := op.MakeOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	p, err := parseConnection(":out", o1)
	assertNoError(t, err)

	if p != o1.OutPort() {
		t.Error("wrong port")
	}
}

func TestParseConnection__SingleIn(t *testing.T) {
	o1, _ := op.MakeOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	o2, _ := op.MakeOperator("o2", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, o1)
	p, err := parseConnection("o2:in", o1)
	assertNoError(t, err)

	if p != o2.InPort() {
		t.Error("wrong port")
	}
}

func TestParseConnection__SingleOut(t *testing.T) {
	o1, _ := op.MakeOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	o2, _ := op.MakeOperator("o2", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, o1)
	p, err := parseConnection("o2:out", o1)
	assertNoError(t, err)

	if p != o2.OutPort() {
		t.Error("wrong port")
	}
}

func TestParseConnection__Map(t *testing.T) {
	o1, _ := op.MakeOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	o2, _ := op.MakeOperator("o2", nil, op.PortDef{Type: "map", Map: map[string]op.PortDef{"a": {Type: "number"}}}, op.PortDef{Type: "number"}, o1)
	p, err := parseConnection("o2:in.a", o1)
	assertNoError(t, err)

	if p != o2.InPort().Port("a") {
		t.Error("wrong port")
	}
}

func TestParseConnection__Map__UnknownKey(t *testing.T) {
	o1, _ := op.MakeOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	op.MakeOperator("o2", nil, op.PortDef{Type: "map", Map: map[string]op.PortDef{"a": {Type: "number"}}}, op.PortDef{Type: "number"}, o1)
	p, err := parseConnection("o2:in.b", o1)
	assertError(t, err)
	assertNil(t, p)
}

func TestParseConnection__Map__DescendingTooDeep(t *testing.T) {
	o1, _ := op.MakeOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	op.MakeOperator("o2", nil, op.PortDef{Type: "map", Map: map[string]op.PortDef{"a": {Type: "number"}}}, op.PortDef{Type: "number"}, o1)
	p, err := parseConnection("o2:in.b.c", o1)
	assertError(t, err)
	assertNil(t, p)
}

func TestParseConnection__NestedMap(t *testing.T) {
	o1, _ := op.MakeOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	o2, _ := op.MakeOperator("o2", nil, op.PortDef{Type: "map", Map: map[string]op.PortDef{"a": {Type: "map", Map: map[string]op.PortDef{"b": {Type: "number"}}}}}, op.PortDef{Type: "number"}, o1)
	p, err := parseConnection("o2:in.a.b", o1)
	assertNoError(t, err)

	if p != o2.InPort().Port("a").Port("b") {
		t.Error("wrong port")
	}
}

func TestParseConnection__Stream(t *testing.T) {
	o1, _ := op.MakeOperator("o1", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	o2, _ := op.MakeOperator("o2", nil, op.PortDef{Type: "stream", Stream: &op.PortDef{Type: "number"}}, op.PortDef{Type: "number"}, o1)
	p, err := parseConnection("o2:in", o1)
	assertNoError(t, err)

	if p != o2.InPort().Stream() {
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
	p, err := parseConnection("o2:in.a.a", o1)
	assertNoError(t, err)

	if p != o2.InPort().Stream().Port("a").Stream().Port("a").Stream() {
		t.Error("wrong port")
	}
}
