package tests

import (
	"slang/op"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOperator_MakeOperator_CorrectRelation(t *testing.T) {
	defPort := op.ParsePortDef(`{"type":"number"}`)
	oParent, _ := op.MakeOperator("parent", nil, defPort, defPort, nil)
	oChild1, _ := op.MakeOperator("child1", nil, defPort, defPort, oParent)
	oChild2, _ := op.MakeOperator("child2", nil, defPort, defPort, oParent)

	if oParent != oChild1.Parent() || oParent != oChild2.Parent() {
		t.Error("oParent must be parent of oChild1 and oChil2")
	}
	if oParent.Child(oChild1.Name()) == nil || oParent.Child(oChild2.Name()) == nil {
		t.Error("oChild1 and oChil2 must be children of oParent")
	}
}

func TestInstanceDef_Validate_Fails_MissingName(t *testing.T) {
	_, err := validateJSONInstanceDef(`{
		"operator": "opr"
	}`)
	assert.Error(t, err)
}

func TestInstanceDef_Validate_Fails_SpacesInName(t *testing.T) {
	_, err := validateJSONInstanceDef(`{
		"operator": "opr",
		"name":"fun 4 ever",
	}`)
	assert.Error(t, err)
}

func TestInstanceDef_Validate_Fails_MissingOperator(t *testing.T) {
	_, err := validateJSONInstanceDef(`{
		"name":"oprInstance"
	}`)
	assert.Error(t, err)
}

func TestInstanceDef_Validate_Succeeds(t *testing.T) {
	ins, err := validateJSONInstanceDef(`{
		"operator": "opr",
		"name":"oprInstance"
	}`)
	assert.NoError(t, err)
	assert.True(t, ins.Valid())
}

func TestOperatorDef_Validate_Fails_PortMustBeDefined_In(t *testing.T) {
	_, err := validateJSONOperatorDef(`{
		"name":"opr",
		"out": {"type":"number"},
	}`)
	assert.Error(t, err)
}

func TestOperatorDef_Validate_Fails_PortMustBeDefined_Out(t *testing.T) {
	_, err := validateJSONOperatorDef(`{
		"name":"opr",
		"in": {"type":"number"},
	}`)
	assert.Error(t, err)
}

func TestOperatorDef_Validate_Succeeds(t *testing.T) {
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
	assert.NoError(t, err)
	assert.True(t, oDef.Valid())
}

func TestOperator_Compile__Nested_1_Child(t *testing.T) {
	op1, _ := op.MakeOperator("", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	op2, _ := op.MakeOperator("a", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, op1)
	op3, _ := op.MakeOperator("b", func(_, _ *op.Port, _ interface{}) {}, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, op2)

	// op1
	op1.In().Connect(op2.In())
	op2.Out().Connect(op1.Out())

	// op2
	op2.In().Connect(op3.In())
	op3.Out().Connect(op2.Out())

	// Compile
	assert.True(t, op1.Compile() == 1)

	assert.True(t, len(op1.Children()) == 1)

	if _, ok := op1.Children()["a.b"]; !ok {
		t.Error("child not there")
	}

	assert.True(t, op3.Parent() == op1)

	assert.True(t, op1.In().Connected(op3.In()))
	assert.True(t, op3.Out().Connected(op1.Out()))

	assert.False(t, op1.In().Connected(op2.In()))
	assert.False(t, op2.Out().Connected(op1.Out()))
}

func TestOperator_Compile__Nested_Children(t *testing.T) {
	op1, _ := op.MakeOperator("", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, nil)
	op2, _ := op.MakeOperator("a", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, op1)
	op3, _ := op.MakeOperator("b", nil, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, op1)
	op4, _ := op.MakeOperator("c", func(_, _ *op.Port, _ interface{}) {}, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, op2)
	op5, _ := op.MakeOperator("d", func(_, _ *op.Port, _ interface{}) {}, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, op2)
	op6, _ := op.MakeOperator("e", func(_, _ *op.Port, _ interface{}) {}, op.PortDef{Type: "number"}, op.PortDef{Type: "number"}, op3)

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
	assert.True(t, op1.Compile() == 2)

	assert.True(t, len(op1.Children()) == 3)

	if _, ok := op1.Children()["a.c"]; !ok {
		t.Error("child not there")
	}

	if _, ok := op1.Children()["a.d"]; !ok {
		t.Error("child not there")
	}

	if _, ok := op1.Children()["b.e"]; !ok {
		t.Error("child not there")
	}

	assert.True(t, op4.Parent() == op1)
	assert.True(t, op5.Parent() == op1)
	assert.True(t, op6.Parent() == op1)

	assert.True(t, op1.In().Connected(op4.In()))
	assert.True(t, op4.Out().Connected(op5.In()))
	assert.True(t, op5.Out().Connected(op6.In()))
	assert.True(t, op6.Out().Connected(op1.Out()))

	assert.False(t, op1.In().Connected(op2.In()))
	assert.False(t, op3.Out().Connected(op1.Out()))
	assert.False(t, op2.In().Connected(op4.In()))
	assert.False(t, op5.Out().Connected(op2.Out()))
	assert.False(t, op3.In().Connected(op6.In()))
	assert.False(t, op6.Out().Connected(op3.Out()))
}
