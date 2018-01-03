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
