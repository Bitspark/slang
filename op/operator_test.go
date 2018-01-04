package op

import (
	"encoding/json"
	"testing"
)

func getJSONOperatorDef(jsonDef string) *OperatorDef {
	def := &OperatorDef{}
	json.Unmarshal([]byte(jsonDef), def)
	return def
}

func validateJSONOperatorDef(jsonDef string) (*OperatorDef, error) {
	def := getJSONOperatorDef(jsonDef)
	return def, def.Validate()
}

func validateJSONInstanceDef(jsonDef string) (*InstanceDef, error) {
	def := &InstanceDef{}
	json.Unmarshal([]byte(jsonDef), def)
	return def, def.Validate()
}

func TestOperator_MakeOperator_CorrectRelation(t *testing.T) {
	defPort := helperJson2PortDef(`{"type":"number"}`)
	oParent, _ := MakeOperator("parent", nil, defPort, defPort, nil)
	oChild1, _ := MakeOperator("child1", nil, defPort, defPort, oParent)
	oChild2, _ := MakeOperator("child2", nil, defPort, defPort, oParent)

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
	assertError(t, err)
}

func TestInstanceDef_Validate_Fails_SpacesInName(t *testing.T) {
	_, err := validateJSONInstanceDef(`{
		"operator": "opr",
		"name":"fun 4 ever",
	}`)
	assertError(t, err)
}

func TestInstanceDef_Validate_Fails_MissingOperator(t *testing.T) {
	_, err := validateJSONInstanceDef(`{
		"name":"oprInstance"
	}`)
	assertError(t, err)
}

func TestInstanceDef_Validate_Succeeds(t *testing.T) {
	ins, err := validateJSONInstanceDef(`{
		"operator": "opr",
		"name":"oprInstance"
	}`)
	assertNoError(t, err)
	assertTrue(t, ins.valid)
}

func TestOperatorDef_Validate_Fails_MissingName(t *testing.T) {
	_, err := validateJSONOperatorDef(`{
		"in": {"type":"number"},
		"out": {"type":"number"},
		"connections": {
			":in": [":out"]
		}
	}`)
	assertError(t, err)
}

func TestOperatorDef_Validate_Fails_SpacesInName(t *testing.T) {
	_, err := validateJSONOperatorDef(`{
		"name":"fun 4 ever",
		"in": {"type":"number"},
		"out": {"type":"number"},
		"connections": {
			":in": [":out"]
		}
	}`)
	assertError(t, err)
}

func TestOperatorDef_Validate_Fails_PortMustBeDefined_In(t *testing.T) {
	_, err := validateJSONOperatorDef(`{
		"name":"opr",
		"out": {"type":"number"},
	}`)
	assertError(t, err)
}

func TestOperatorDef_Validate_Fails_PortMustBeDefined_Out(t *testing.T) {
	_, err := validateJSONOperatorDef(`{
		"name":"opr",
		"in": {"type":"number"},
	}`)
	assertError(t, err)
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
	assertNoError(t, err)
	assertTrue(t, oDef.valid)
}

func TestOperator_Compile__Nested_1_Child(t *testing.T) {
	op1, _ := MakeOperator("", nil, PortDef{Type: "number"}, PortDef{Type: "number"}, nil)
	op2, _ := MakeOperator("a", nil, PortDef{Type: "number"}, PortDef{Type: "number"}, op1)
	op3, _ := MakeOperator("b", func(_, _ *Port, _ interface{}) {}, PortDef{Type: "number"}, PortDef{Type: "number"}, op2)

	// op1
	op1.In().Connect(op2.In())
	op2.Out().Connect(op1.Out())

	// op2
	op2.In().Connect(op3.In())
	op3.Out().Connect(op2.Out())

	// Compile
	assertTrue(t, op1.Compile() == 1)

	assertTrue(t, len(op1.children) == 1)

	if _, ok := op1.children["a.b"]; !ok {
		t.Error("child not there")
	}

	assertTrue(t, op3.parent == op1)

	assertTrue(t, op1.In().Connected(op3.In()))
	assertTrue(t, op3.Out().Connected(op1.Out()))

	assertFalse(t, op1.In().Connected(op2.In()))
	assertFalse(t, op2.Out().Connected(op1.Out()))
}

func TestOperator_Compile__Nested_Children(t *testing.T) {
	op1, _ := MakeOperator("", nil, PortDef{Type: "number"}, PortDef{Type: "number"}, nil)
	op2, _ := MakeOperator("a", nil, PortDef{Type: "number"}, PortDef{Type: "number"}, op1)
	op3, _ := MakeOperator("b", nil, PortDef{Type: "number"}, PortDef{Type: "number"}, op1)
	op4, _ := MakeOperator("c", func(_, _ *Port, _ interface{}) {}, PortDef{Type: "number"}, PortDef{Type: "number"}, op2)
	op5, _ := MakeOperator("d", func(_, _ *Port, _ interface{}) {}, PortDef{Type: "number"}, PortDef{Type: "number"}, op2)
	op6, _ := MakeOperator("e", func(_, _ *Port, _ interface{}) {}, PortDef{Type: "number"}, PortDef{Type: "number"}, op3)

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
	assertTrue(t, op1.Compile() == 2)

	assertTrue(t, len(op1.children) == 3)

	if _, ok := op1.children["a.c"]; !ok {
		t.Error("child not there")
	}

	if _, ok := op1.children["a.d"]; !ok {
		t.Error("child not there")
	}

	if _, ok := op1.children["b.e"]; !ok {
		t.Error("child not there")
	}

	assertTrue(t, op4.parent == op1)
	assertTrue(t, op5.parent == op1)
	assertTrue(t, op6.parent == op1)

	assertTrue(t, op1.In().Connected(op4.In()))
	assertTrue(t, op4.Out().Connected(op5.In()))
	assertTrue(t, op5.Out().Connected(op6.In()))
	assertTrue(t, op6.Out().Connected(op1.Out()))

	assertFalse(t, op1.In().Connected(op2.In()))
	assertFalse(t, op3.Out().Connected(op1.Out()))
	assertFalse(t, op2.In().Connected(op4.In()))
	assertFalse(t, op5.Out().Connected(op2.Out()))
	assertFalse(t, op3.In().Connected(op6.In()))
	assertFalse(t, op6.Out().Connected(op3.Out()))
}
