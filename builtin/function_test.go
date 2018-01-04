package builtin

import (
	"slang/op"
	"testing"
)

func TestManager_MakeOperator__Function__NilProperties(t *testing.T) {
	fo, err := MakeOperator(op.InstanceDef{Operator: "function"}, nil)

	if fo != nil || err == nil {
		t.Error("expected error")
	}
}

func TestManager_MakeOperator__Function__EmptyExpression(t *testing.T) {
	fo, err := MakeOperator(op.InstanceDef{Operator: "function", Properties: map[string]interface{}{"expression": ""}}, nil)

	if fo != nil || err == nil {
		t.Error("expected error")
	}
}

func TestManager_MakeOperator__Function__InvalidExpression(t *testing.T) {
	fo, err := MakeOperator(op.InstanceDef{Operator: "function", Properties: map[string]interface{}{"expression": "+"}}, nil)

	if fo != nil || err == nil {
		t.Error("expected error")
	}
}

func TestManager_MakeOperator__Function__Add(t *testing.T) {
	fo, err := MakeOperator(op.InstanceDef{Operator: "function", Properties: map[string]interface{}{"expression": "a+b"}}, nil)

	if fo == nil {
		t.Error("operator not defined")
	}

	if err != nil {
		t.Error(err.Error())
	}

	if fo.In().Type() != op.TYPE_MAP {
		t.Error("expected map")
	}

	if fo.In().Map("a").Type() != op.TYPE_ANY {
		t.Error("expected any input")
	}

	if fo.In().Map("b").Type() != op.TYPE_ANY {
		t.Error("expected any input")
	}

	fo.Out().Bufferize()

	go fo.Start()

	fo.In().Push(map[string]interface{}{"a": 1.0, "b": 2.0})
	fo.In().Push(map[string]interface{}{"a": -5.0, "b": 2.5})
	fo.In().Push(map[string]interface{}{"a": 0.0, "b": 333.0})

	if fo.Out().Pull() != 3.0 {
		t.Error("wrong output")
	}

	if fo.Out().Pull() != -2.5 {
		t.Error("wrong output")
	}

	if fo.Out().Pull() != 333.0 {
		t.Error("wrong output")
	}
}

func TestManager_MakeOperator__Function__BoolArith(t *testing.T) {
	fo, err := MakeOperator(op.InstanceDef{Operator: "function", Properties: map[string]interface{}{"expression": "a && (b != c)"}}, nil)

	if fo == nil {
		t.Error("operator not defined")
	}

	if err != nil {
		t.Error(err.Error())
	}

	if fo.In().Type() != op.TYPE_MAP {
		t.Error("expected map")
	}

	if fo.In().Map("a").Type() != op.TYPE_ANY {
		t.Error("expected any input")
	}

	if fo.In().Map("b").Type() != op.TYPE_ANY {
		t.Error("expected any input")
	}

	if fo.In().Map("c").Type() != op.TYPE_ANY {
		t.Error("expected any input")
	}

	fo.Out().Bufferize()

	go fo.Start()

	fo.In().Push(map[string]interface{}{"a": true, "b": true, "c": false})
	fo.In().Push(map[string]interface{}{"a": false, "b": false, "c": false})
	fo.In().Push(map[string]interface{}{"a": false, "b": false, "c": true})
	fo.In().Push(map[string]interface{}{"a": true, "b": false, "c": true})
	fo.In().Push(map[string]interface{}{"a": true, "b": false, "c": false})

	if fo.Out().Pull() != true {
		t.Error("wrong output")
	}

	if fo.Out().Pull() != false {
		t.Error("wrong output")
	}

	if fo.Out().Pull() != false {
		t.Error("wrong output")
	}

	if fo.Out().Pull() != true {
		t.Error("wrong output")
	}

	if fo.Out().Pull() != false {
		t.Error("wrong output")
	}
}
