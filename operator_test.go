package slang

import (
	"testing"
)

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
