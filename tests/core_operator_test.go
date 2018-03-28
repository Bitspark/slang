package tests

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"testing"
	"github.com/Bitspark/slang/pkg/api"
)

func TestOperator_NewOperator__CorrectRelation(t *testing.T) {
	defPort := api.ParsePortDef(`{"type":"number"}`)
	oParent, _ := core.NewOperator("parent", nil, nil, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: defPort, Out: defPort}}, nil)
	oChild1, _ := core.NewOperator("child1", nil, nil, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: defPort, Out: defPort}}, nil)
	oChild1.SetParent(oParent)
	oChild2, _ := core.NewOperator("child2", nil, nil, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: defPort, Out: defPort}}, nil)
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
	op1, _ := core.NewOperator("", nil, nil, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: core.TypeDef{Type: "number"}, Out: core.TypeDef{Type: "number"}}}, nil)
	op2, _ := core.NewOperator("a", nil, nil, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: core.TypeDef{Type: "number"}, Out: core.TypeDef{Type: "number"}}}, nil)
	op2.SetParent(op1)
	op3, _ := core.NewOperator("b", func(*core.Operator) {}, nil, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: core.TypeDef{Type: "number"}, Out: core.TypeDef{Type: "number"}}}, nil)
	op3.SetParent(op2)

	// op1
	op1.Main().In().Connect(op2.Main().In())
	op2.Main().Out().Connect(op1.Main().Out())

	// op2
	op2.Main().In().Connect(op3.Main().In())
	op3.Main().Out().Connect(op2.Main().Out())

	// Compile
	a.Equal(1, op1.Compile())

	a.True(len(op1.Children()) == 1)

	if _, ok := op1.Children()["a.b"]; !ok {
		t.Error("child not there")
	}

	a.True(op3.Parent() == op1)

	a.True(op1.Main().In().Connected(op3.Main().In()))
	a.True(op3.Main().Out().Connected(op1.Main().Out()))

	a.False(op1.Main().In().Connected(op2.Main().In()))
	a.False(op2.Main().Out().Connected(op1.Main().Out()))
}

func TestOperator_Compile__NestedChildren(t *testing.T) {
	a := assertions.New(t)
	op1, _ := core.NewOperator("", nil, nil, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: core.TypeDef{Type: "number"}, Out: core.TypeDef{Type: "number"}}}, nil)
	op2, _ := core.NewOperator("a", nil, nil, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: core.TypeDef{Type: "number"}, Out: core.TypeDef{Type: "number"}}}, nil)
	op2.SetParent(op1)
	op3, _ := core.NewOperator("b", nil, nil, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: core.TypeDef{Type: "number"}, Out: core.TypeDef{Type: "number"}}}, nil)
	op3.SetParent(op1)
	op4, _ := core.NewOperator("c", func(*core.Operator) {}, nil, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: core.TypeDef{Type: "number"}, Out: core.TypeDef{Type: "number"}}}, nil)
	op4.SetParent(op2)
	op5, _ := core.NewOperator("d", func(*core.Operator) {}, nil, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: core.TypeDef{Type: "number"}, Out: core.TypeDef{Type: "number"}}}, nil)
	op5.SetParent(op2)
	op6, _ := core.NewOperator("e", func(*core.Operator) {}, nil, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: core.TypeDef{Type: "number"}, Out: core.TypeDef{Type: "number"}}}, nil)
	op6.SetParent(op3)

	// op1
	op1.Main().In().Connect(op2.Main().In())
	op2.Main().Out().Connect(op3.Main().In())
	op3.Main().Out().Connect(op1.Main().Out())

	// op2
	op2.Main().In().Connect(op4.Main().In())
	op4.Main().Out().Connect(op5.Main().In())
	op5.Main().Out().Connect(op2.Main().Out())

	// op3
	op3.Main().In().Connect(op6.Main().In())
	op6.Main().Out().Connect(op3.Main().Out())

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

	a.True(op1.Main().In().Connected(op4.Main().In()))
	a.True(op4.Main().Out().Connected(op5.Main().In()))
	a.True(op5.Main().Out().Connected(op6.Main().In()))
	a.True(op6.Main().Out().Connected(op1.Main().Out()))

	a.False(op1.Main().In().Connected(op2.Main().In()))
	a.False(op3.Main().Out().Connected(op1.Main().Out()))
	a.False(op2.Main().In().Connected(op4.Main().In()))
	a.False(op5.Main().Out().Connected(op2.Main().Out()))
	a.False(op3.Main().In().Connected(op6.Main().In()))
	a.False(op6.Main().Out().Connected(op3.Main().Out()))
}
