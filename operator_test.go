package slang

import (
	"encoding/json"
	"testing"
)

func validateJsonOperatorDef(jsonDef string) (*OperatorDef, error) {
	def := &OperatorDef{}
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

func TestOperatorDef_Validate_Fails_NoSpacesInOperatorName(t *testing.T) {
	_, err := validateJsonOperatorDef(`{
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
	_, err := validateJsonOperatorDef(`{
		"name":"opr",
		"out": {"type":"number"},
	}`)
	assertError(t, err)
}

func TestOperatorDef_Validate_Fails_PortMustBeDefined_Out(t *testing.T) {
	_, err := validateJsonOperatorDef(`{
		"name":"opr",
		"in": {"type":"number"},
	}`)
	assertError(t, err)
}

func TestOperatorDef_Validate_Succeeds(t *testing.T) {
	oDef, err := validateJsonOperatorDef(`{
		"name": "opr",
		"in": {
			"type": "number"
		},
		"out": {
			"type": "number"
		},
		"operators": {
			"add": {
				"class": "dummyBuildIn"
			}
		},
		"connections": {
			"opr.in": ["add.in"],
			"add.out": ["opr.in"]
		}
	}`)
	assertNoError(t, err)
	assertTrue(t, oDef.valid)
}

func TestOperator_ParseOperator_TrivialCase(t *testing.T) {
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
}

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
