package tests

import (
	"slang"
	"slang/core"
	"slang/tests/assertions"
	"testing"
)

func TestOperator_NewOperator_CorrectRelation(t *testing.T) {
	defPort := slang.ParsePortDef(`{"type":"number"}`)
	oParent, _ := core.NewOperator("parent", nil, defPort, defPort, nil)
	oChild1, _ := core.NewOperator("child1", nil, defPort, defPort, oParent)
	oChild2, _ := core.NewOperator("child2", nil, defPort, defPort, oParent)

	if oParent != oChild1.Parent() || oParent != oChild2.Parent() {
		t.Error("oParent must be parent of oChild1 and oChil2")
	}
	if oParent.Child(oChild1.Name()) == nil || oParent.Child(oChild2.Name()) == nil {
		t.Error("oChild1 and oChil2 must be children of oParent")
	}
}

func TestInstanceDef_Validate_Fails_MissingName(t *testing.T) {
	a := assertions.New(t)
	_, err := validateJSONInstanceDef(`{
		"operator": "opr"
	}`)
	a.Error(err)
}

func TestInstanceDef_Validate_Fails_SpacesInName(t *testing.T) {
	a := assertions.New(t)
	_, err := validateJSONInstanceDef(`{
		"operator": "opr",
		"name":"fun 4 ever",
	}`)
	a.Error(err)
}

func TestInstanceDef_Validate_Fails_MissingOperator(t *testing.T) {
	a := assertions.New(t)
	_, err := validateJSONInstanceDef(`{
		"name":"oprInstance"
	}`)
	a.Error(err)
}

func TestInstanceDef_Validate_Succeeds(t *testing.T) {
	a := assertions.New(t)
	ins, err := validateJSONInstanceDef(`{
		"operator": "opr",
		"name":"oprInstance"
	}`)
	a.NoError(err)
	a.True(ins.Valid())
}

func TestOperatorDef_Validate_Fails_PortMustBeDefined_In(t *testing.T) {
	a := assertions.New(t)
	_, err := validateJSONOperatorDef(`{
		"name":"opr",
		"out": {"type":"number"},
	}`)
	a.Error(err)
}

func TestOperatorDef_Validate_Fails_PortMustBeDefined_Out(t *testing.T) {
	a := assertions.New(t)
	_, err := validateJSONOperatorDef(`{
		"name":"opr",
		"in": {"type":"number"},
	}`)
	a.Error(err)
}

func TestOperatorDef_Validate_Succeeds(t *testing.T) {
	a := assertions.New(t)
	oDef, err := validateJSONOperatorDef(`{
		"name": "opr",
		"in": {
			"type": "number"
		},
		"out": {
			"type": "number"
		},
		"operators": [
			{
				"operator": "builtin_Adder",
				"name": "add"
			}
		],
		"connections": {
			":in": ["add:in"],
			"add:out": [":in"]
		}
	}`)
	a.NoError(err)
	a.True(oDef.Valid())
}

func TestOperator_Compile__Nested_1_Child(t *testing.T) {
	a := assertions.New(t)
	op1, _ := core.NewOperator("", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	op2, _ := core.NewOperator("a", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, op1)
	op3, _ := core.NewOperator("b", func(_, _ *core.Port, _ interface{}) {}, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, op2)

	// op1
	op1.In().Connect(op2.In())
	op2.Out().Connect(op1.Out())

	// op2
	op2.In().Connect(op3.In())
	op3.Out().Connect(op2.Out())

	// Compile
	a.True(op1.Compile() == 1)

	a.True(len(op1.Children()) == 1)

	if _, ok := op1.Children()["a.b"]; !ok {
		t.Error("child not there")
	}

	a.True(op3.Parent() == op1)

	a.True(op1.In().Connected(op3.In()))
	a.True(op3.Out().Connected(op1.Out()))

	a.False(op1.In().Connected(op2.In()))
	a.False(op2.Out().Connected(op1.Out()))
}

func TestOperator_Compile__Nested_Children(t *testing.T) {
	a := assertions.New(t)
	op1, _ := core.NewOperator("", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	op2, _ := core.NewOperator("a", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, op1)
	op3, _ := core.NewOperator("b", nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, op1)
	op4, _ := core.NewOperator("c", func(_, _ *core.Port, _ interface{}) {}, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, op2)
	op5, _ := core.NewOperator("d", func(_, _ *core.Port, _ interface{}) {}, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, op2)
	op6, _ := core.NewOperator("e", func(_, _ *core.Port, _ interface{}) {}, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, op3)

	// op1
	op1.In().Connect(op2.In())
	op2.Out().Connect(op3.In())
	op3.Out().Connect(op1.Out())

	// op2
	op2.In().Connect(op4.In())
	op4.Out().Connect(op5.In())
	op5.Out().Connect(op2.Out())

	// op3
	op3.In().Connect(op6.In())
	op6.Out().Connect(op3.Out())

	// Compile
	a.True(op1.Compile() == 2)

	a.True(len(op1.Children()) == 3)

	if _, ok := op1.Children()["a.c"]; !ok {
		t.Error("child not there")
	}

	if _, ok := op1.Children()["a.d"]; !ok {
		t.Error("child not there")
	}

	if _, ok := op1.Children()["b.e"]; !ok {
		t.Error("child not there")
	}

	a.True(op4.Parent() == op1)
	a.True(op5.Parent() == op1)
	a.True(op6.Parent() == op1)

	a.True(op1.In().Connected(op4.In()))
	a.True(op4.Out().Connected(op5.In()))
	a.True(op5.Out().Connected(op6.In()))
	a.True(op6.Out().Connected(op1.Out()))

	a.False(op1.In().Connected(op2.In()))
	a.False(op3.Out().Connected(op1.Out()))
	a.False(op2.In().Connected(op4.In()))
	a.False(op5.Out().Connected(op2.Out()))
	a.False(op3.In().Connected(op6.In()))
	a.False(op6.Out().Connected(op3.Out()))
}

func TestOperator_SpecifyAny__Simple(t *testing.T) {
	a := assertions.New(t)

	in := core.PortDef{Type: "any", Any: "ident1"}
	out := core.PortDef{Type: "string"}
	o, err := core.NewOperator("any1", nil, in, out, nil)
	a.NoError(err)
	a.NotNil(o)

	err = o.SpecifyAny("ident1", core.PortDef{Type:"number"})
	a.NoError(err)

	a.Equal(core.TYPE_NUMBER, o.In().Type())
	a.True(o.AnysSpecified())
}

func TestOperator_SpecifyAny__UnknownIdentifier(t *testing.T) {
	a := assertions.New(t)

	in := core.PortDef{Type: "any", Any: "ident1"}
	out := core.PortDef{Type: "string"}
	o, _ := core.NewOperator("any1", nil, in, out, nil)

	err := o.SpecifyAny("ident2", core.PortDef{Type:"number"})
	a.Error(err)
}

func TestOperator_SpecifyAny__DoubleDefinition(t *testing.T) {
	a := assertions.New(t)

	in := core.PortDef{Type: "any", Any: "ident1"}
	out := core.PortDef{Type: "string"}
	o, _ := core.NewOperator("any1", nil, in, out, nil)

	err := o.SpecifyAny("ident1", core.PortDef{Type:"number"})
	a.NoError(err)
	err = o.SpecifyAny("ident1", core.PortDef{Type:"string"})
	a.Error(err)
}

func TestOperator_SpecifyAny__ReplacingAll(t *testing.T) {
	a := assertions.New(t)

	in := core.PortDef{Type: "any", Any: "ident1"}
	out := core.PortDef{Type: "any", Any: "ident1"}
	o, _ := core.NewOperator("any1", nil, in, out, nil)

	o.SpecifyAny("ident1", core.PortDef{Type:"string"})

	a.Equal(core.TYPE_STRING, o.In().Type())
	a.Equal(core.TYPE_STRING, o.Out().Type())
	a.True(o.AnysSpecified())
}

func TestOperator_SpecifyAny__NotReplacingOthers(t *testing.T) {
	a := assertions.New(t)

	in := core.PortDef{Type: "any", Any: "ident1"}
	out := core.PortDef{Type: "any", Any: "ident2"}
	o, _ := core.NewOperator("any1", nil, in, out, nil)

	o.SpecifyAny("ident1", core.PortDef{Type:"boolean"})

	a.Equal(core.TYPE_BOOLEAN, o.In().Type())
	a.Equal(core.TYPE_ANY, o.Out().Type())
	a.False(o.AnysSpecified())
}

func TestOperator_ConnectAndReplace(t *testing.T) {
	a := assertions.New(t)

	in := core.PortDef{Type: "number"}
	out := core.PortDef{Type: "number"}
	o1, _ := core.NewOperator("any1", nil, in, out, nil)

	in = core.PortDef{Type: "any", Any: "ident1"}
	out = core.PortDef{Type: "number"}
	o2, _ := core.NewOperator("any1", nil, in, out, nil)
	a.Equal(core.TYPE_ANY, o2.In().Type())

	a.NoError(o1.Out().Connect(o2.In()))
	a.Equal(core.TYPE_NUMBER, o2.In().Type())
}

func TestOperator_ConnectCollidingTypes(t *testing.T) {
	a := assertions.New(t)

	in := core.PortDef{Type: "number"}
	out := core.PortDef{Type: "number"}
	o1, _ := core.NewOperator("num", nil, in, out, nil)

	in = core.PortDef{Type: "any", Any: "ident1"}
	out = core.PortDef{Type: "any", Any: "ident1"}
	o2, _ := core.NewOperator("any", nil, in, out, nil)

	in = core.PortDef{Type: "boolean"}
	out = core.PortDef{Type: "boolean"}
	o3, _ := core.NewOperator("bool", nil, in, out, nil)

	a.NoError(o2.Out().Connect(o3.In()))
	a.Error(o1.Out().Connect(o2.In()))
}

func TestOperator_ConnectNoCollidingTypes(t *testing.T) {
	a := assertions.New(t)

	in := core.PortDef{Type: "number"}
	out := core.PortDef{Type: "number"}
	o1, _ := core.NewOperator("num", nil, in, out, nil)

	in = core.PortDef{Type: "any", Any: "ident1"}
	out = core.PortDef{Type: "any", Any: "ident2"}
	o2, _ := core.NewOperator("any", nil, in, out, nil)

	in = core.PortDef{Type: "boolean"}
	out = core.PortDef{Type: "boolean"}
	o3, _ := core.NewOperator("bool", nil, in, out, nil)

	a.NoError(o2.Out().Connect(o3.In()))
	a.NoError(o1.Out().Connect(o2.In()))
}

func TestOperator_ConnectInternally(t *testing.T) {
	a := assertions.New(t)

	in := core.PortDef{Type: "number"}
	out := core.PortDef{Type: "any", Any: "bla"}
	o1, _ := core.NewOperator("any1", nil, in, out, nil)
	a.Equal(core.TYPE_ANY, o1.Out().Type())

	a.NoError(o1.In().Connect(o1.Out()))
	a.Equal(core.TYPE_NUMBER, o1.Out().Type())
}

func TestOperator_ConnectNested(t *testing.T) {
	a := assertions.New(t)

	in := core.PortDef{Type: "number"}
	out := core.PortDef{Type: "any", Any: "a"}
	o1, _ := core.NewOperator("any1", nil, in, out, nil)
	in = core.PortDef{Type: "any", Any: "b"}
	out = core.PortDef{Type: "any", Any: "b"}
	o2, _ := core.NewOperator("any2", nil, in, out, o1)
	a.Equal(core.TYPE_ANY, o1.Out().Type())
	a.Equal(core.TYPE_ANY, o2.In().Type())
	a.Equal(core.TYPE_ANY, o2.Out().Type())

	a.NoError(o2.In().Connect(o2.Out()))
	a.NoError(o2.Out().Connect(o1.Out()))
	a.NoError(o1.In().Connect(o2.In()))

	a.Equal(core.TYPE_NUMBER, o2.In().Type())
	a.Equal(core.TYPE_NUMBER, o2.Out().Type())
	a.Equal(core.TYPE_NUMBER, o1.Out().Type())
}

func TestOperator_ConnectUltraNested(t *testing.T) {
	a := assertions.New(t)

	in := core.PortDef{Type: "number"}
	out := core.PortDef{Type: "any", Any: "a"}
	o1, _ := core.NewOperator("any1", nil, in, out, nil)
	in = core.PortDef{Type: "any", Any: "b"}
	out = core.PortDef{Type: "any", Any: "b"}
	o2, _ := core.NewOperator("any1", nil, in, out, o1)
	in = core.PortDef{Type: "any", Any: "c"}
	out = core.PortDef{Type: "any", Any: "c"}
	o3, _ := core.NewOperator("any1", nil, in, out, o2)
	a.Equal(core.TYPE_ANY, o1.Out().Type())
	a.Equal(core.TYPE_ANY, o2.In().Type())
	a.Equal(core.TYPE_ANY, o2.Out().Type())
	a.Equal(core.TYPE_ANY, o3.In().Type())
	a.Equal(core.TYPE_ANY, o3.Out().Type())

	a.NoError(o3.In().Connect(o3.Out()))

	a.NoError(o2.In().Connect(o3.In()))
	a.NoError(o3.Out().Connect(o2.Out()))

	a.NoError(o1.In().Connect(o2.In()))
	a.NoError(o2.Out().Connect(o1.Out()))

	a.Equal(core.TYPE_NUMBER, o1.Out().Type())
	a.Equal(core.TYPE_NUMBER, o2.In().Type())
	a.Equal(core.TYPE_NUMBER, o2.Out().Type())
	a.Equal(core.TYPE_NUMBER, o3.In().Type())
	a.Equal(core.TYPE_NUMBER, o3.Out().Type())
}