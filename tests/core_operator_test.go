package tests

import (
	"slang"
	"slang/core"
	"slang/tests/assertions"
	"testing"
)

func TestOperator_NewOperator__CorrectRelation(t *testing.T) {
	defPort := slang.ParsePortDef(`{"type":"number"}`)
	oParent, _ := core.NewOperator("parent", nil, nil, defPort, defPort, nil)
	oChild1, _ := core.NewOperator("child1", nil, nil, defPort, defPort, nil)
	oChild1.SetParent(oParent)
	oChild2, _ := core.NewOperator("child2", nil, nil, defPort, defPort, nil)
	oChild2.SetParent(oParent)

	if oParent != oChild1.Parent() || oParent != oChild2.Parent() {
		t.Error("oParent must be parent of oChild1 and oChil2")
	}
	if oParent.Child(oChild1.Name()) == nil || oParent.Child(oChild2.Name()) == nil {
		t.Error("oChild1 and oChil2 must be children of oParent")
	}
}

func TestOperator_Compile__Nested1Child(t *testing.T) {
	a := assertions.New(t)
	op1, _ := core.NewOperator("", nil, nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	op2, _ := core.NewOperator("a", nil, nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	op2.SetParent(op1)
	op3, _ := core.NewOperator("b", func(_, _ *core.Port, _ map[string]*core.Delegate, _ interface{}) {}, nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	op3.SetParent(op2)

	// op1
	op1.In().Connect(op2.In())
	op2.Out().Connect(op1.Out())

	// op2
	op2.In().Connect(op3.In())
	op3.Out().Connect(op2.Out())

	// Compile
	a.Equal(1, op1.Compile())

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

func TestOperator_Compile__NestedChildren(t *testing.T) {
	a := assertions.New(t)
	op1, _ := core.NewOperator("", nil, nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	op2, _ := core.NewOperator("a", nil, nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	op2.SetParent(op1)
	op3, _ := core.NewOperator("b", nil, nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	op3.SetParent(op1)
	op4, _ := core.NewOperator("c", func(_, _ *core.Port, _ map[string]*core.Delegate, _ interface{}) {}, nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	op4.SetParent(op2)
	op5, _ := core.NewOperator("d", func(_, _ *core.Port, _ map[string]*core.Delegate, _ interface{}) {}, nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	op5.SetParent(op2)
	op6, _ := core.NewOperator("e", func(_, _ *core.Port, _ map[string]*core.Delegate, _ interface{}) {}, nil, core.PortDef{Type: "number"}, core.PortDef{Type: "number"}, nil)
	op6.SetParent(op3)

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
	a.Equal(2, op1.Compile())

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
