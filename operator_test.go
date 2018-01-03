package slang

import (
	"encoding/json"
	"testing"
)

func validateJSONOperatorDef(jsonDef string) (*OperatorDef, error) {
	def := &OperatorDef{}
	json.Unmarshal([]byte(jsonDef), def)
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
	_, err := validateJSONInstanceDef(`{
		"operator": "opr",
		"name":"oprInstance"
	}`)
	assertNoError(t, err)
}

func TestOperatorDef_Validate_Fails_MissingName(t *testing.T) {
	_, err := validateJSONOperatorDef(`{
		"in": {"type":"number"},
		"out": {"type":"number"},
		"connections": {
			"opr.in": ["opr.out"]
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
			"opr.in": ["opr.out"]
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
			"opr.in": ["add.in"],
			"add.out": ["opr.in"]
		}
	}`)
	assertNoError(t, err)
	assertTrue(t, oDef.valid)
}

/*func TestOperator_ParseOperator_TrivialCase(t *testing.T) {
	jsonDef := `{
		"name":"opr",
		"in": {"type":"number"},
		"out": {"type":"number"},
		"connections": {
			"opr.in": ["opr.out"]
		}
	}`
	o, err := ParseOperator(jsonDef)
	assertNoError(t, err)
	o.InPort().Connected(o.OutPort())
}*/

/*
func TestOperator_ParseOperator_OperatorContaining_1_Builtin(t *testing.T) {
	jsonDef := `{
		"name":"fun",
		"in": {"type":"number"},
		"out": {"type":"number"},
		"operators": {
			"class": "anyBuiltIn"
		}
		"connections": {
			"in": [""]
		}
	}`
	o, err := ParseOperator(jsonDef)
	assertNoError(t, err)
}
*/

func TestParseConnection__NilOperator(t *testing.T) {
	p, err := parseConnection("test.in", nil)
	assertError(t, err)
	assertNil(t, p)
}

func TestParseConnection__NilConnection(t *testing.T) {
	o1, _ := MakeOperator("o1", nil, PortDef{Type: "number"}, PortDef{Type: "number"}, nil)
	MakeOperator("o2", nil, PortDef{Type: "number"}, PortDef{Type: "number"}, o1)
	p, err := parseConnection("", o1)
	assertError(t, err)
	assertNil(t, p)
}

func TestParseConnection__SelfIn(t *testing.T) {
	o1, _ := MakeOperator("o1", nil, PortDef{Type: "number"}, PortDef{Type: "number"}, nil)
	p, err := parseConnection(":in", o1)
	assertNoError(t, err)

	if p != o1.InPort() {
		t.Error("wrong port")
	}
}

func TestParseConnection__SelfOut(t *testing.T) {
	o1, _ := MakeOperator("o1", nil, PortDef{Type: "number"}, PortDef{Type: "number"}, nil)
	p, err := parseConnection(":out", o1)
	assertNoError(t, err)

	if p != o1.OutPort() {
		t.Error("wrong port")
	}
}

func TestParseConnection__SingleIn(t *testing.T) {
	o1, _ := MakeOperator("o1", nil, PortDef{Type: "number"}, PortDef{Type: "number"}, nil)
	o2, _ := MakeOperator("o2", nil, PortDef{Type: "number"}, PortDef{Type: "number"}, o1)
	p, err := parseConnection("o2:in", o1)
	assertNoError(t, err)

	if p != o2.InPort() {
		t.Error("wrong port")
	}
}

func TestParseConnection__SingleOut(t *testing.T) {
	o1, _ := MakeOperator("o1", nil, PortDef{Type: "number"}, PortDef{Type: "number"}, nil)
	o2, _ := MakeOperator("o2", nil, PortDef{Type: "number"}, PortDef{Type: "number"}, o1)
	p, err := parseConnection("o2:out", o1)
	assertNoError(t, err)

	if p != o2.OutPort() {
		t.Error("wrong port")
	}
}

func TestParseConnection__Map(t *testing.T) {
	o1, _ := MakeOperator("o1", nil, PortDef{Type: "number"}, PortDef{Type: "number"}, nil)
	o2, _ := MakeOperator("o2", nil, PortDef{Type: "map", Map: map[string]PortDef{"a": {Type: "number"}}}, PortDef{Type: "number"}, o1)
	p, err := parseConnection("o2:in.a", o1)
	assertNoError(t, err)

	if p != o2.InPort().Port("a") {
		t.Error("wrong port")
	}
}

func TestParseConnection__Map__UnknownKey(t *testing.T) {
	o1, _ := MakeOperator("o1", nil, PortDef{Type: "number"}, PortDef{Type: "number"}, nil)
	MakeOperator("o2", nil, PortDef{Type: "map", Map: map[string]PortDef{"a": {Type: "number"}}}, PortDef{Type: "number"}, o1)
	p, err := parseConnection("o2:in.b", o1)
	assertError(t, err)
	assertNil(t, p)
}

func TestParseConnection__Map__DescendingTooDeep(t *testing.T) {
	o1, _ := MakeOperator("o1", nil, PortDef{Type: "number"}, PortDef{Type: "number"}, nil)
	MakeOperator("o2", nil, PortDef{Type: "map", Map: map[string]PortDef{"a": {Type: "number"}}}, PortDef{Type: "number"}, o1)
	p, err := parseConnection("o2:in.b.c", o1)
	assertError(t, err)
	assertNil(t, p)
}

func TestParseConnection__NestedMap(t *testing.T) {
	o1, _ := MakeOperator("o1", nil, PortDef{Type: "number"}, PortDef{Type: "number"}, nil)
	o2, _ := MakeOperator("o2", nil, PortDef{Type: "map", Map: map[string]PortDef{"a": {Type: "map", Map: map[string]PortDef{"b": {Type: "number"}}}}}, PortDef{Type: "number"}, o1)
	p, err := parseConnection("o2:in.a.b", o1)
	assertNoError(t, err)

	if p != o2.InPort().Port("a").Port("b") {
		t.Error("wrong port")
	}
}

func TestParseConnection__Stream(t *testing.T) {
	o1, _ := MakeOperator("o1", nil, PortDef{Type: "number"}, PortDef{Type: "number"}, nil)
	o2, _ := MakeOperator("o2", nil, PortDef{Type: "stream", Stream: &PortDef{Type: "number"}}, PortDef{Type: "number"}, o1)
	p, err := parseConnection("o2:in", o1)
	assertNoError(t, err)

	if p != o2.InPort().Stream() {
		t.Error("wrong port")
	}
}

func TestParseConnection__StreamMap(t *testing.T) {
	o1, _ := MakeOperator("o1", nil, PortDef{Type: "number"}, PortDef{Type: "number"}, nil)
	o2, _ := MakeOperator("o2", nil,
		PortDef{
			Type: "stream",
			Stream: &PortDef{
				Type: "map",
				Map: map[string]PortDef{
					"a": {
						Type: "stream",
						Stream: &PortDef{
							Type: "map",
							Map: map[string]PortDef{
								"a": {
									Type: "stream",
									Stream: &PortDef{
										Type: "boolean",
									},
								},
							},
						},
					},
				}},
		},
		PortDef{Type: "number"}, o1)
	p, err := parseConnection("o2:in.a.a", o1)
	assertNoError(t, err)

	if p != o2.InPort().Stream().Port("a").Stream().Port("a").Stream() {
		t.Error("wrong port")
	}
}
