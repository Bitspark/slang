package builtin

import (
	"slang/core"
	"slang/tests/assertions"
	"testing"
)

func TestBuiltin_Eval__CreatorFuncIsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFunction := getCreatorFunc("eval")
	a.NotNil(ocFunction)
}

func TestBuiltin_Eval__NilProperties(t *testing.T) {
	fo, err := createOpEval(core.InstanceDef{Operator: "eval"}, nil)

	if fo != nil || err == nil {
		t.Error("expected error")
	}
}

func TestBuiltin_Eval__EmptyExpression(t *testing.T) {
	fo, err := createOpEval(core.InstanceDef{Operator: "eval", Properties: map[string]interface{}{"expression": ""}}, nil)

	if fo != nil || err == nil {
		t.Error("expected error")
	}
}

func TestBuiltin_Eval__InvalidExpression(t *testing.T) {
	fo, err := createOpEval(core.InstanceDef{Operator: "eval", Properties: map[string]interface{}{"expression": "+"}}, nil)

	if fo != nil || err == nil {
		t.Error("expected error")
	}
}

func TestBuiltin_Eval__Add(t *testing.T) {
	fo, err := createOpEval(core.InstanceDef{Operator: "eval", Properties: map[string]interface{}{"expression": "a+b"}}, nil)

	if fo == nil {
		t.Error("operator not defined")
	}

	if err != nil {
		t.Error(err.Error())
	}

	if fo.In().Type() != core.TYPE_MAP {
		t.Error("expected map")
	}

	if fo.In().Map("a").Type() != core.TYPE_ANY {
		t.Error("expected any input")
	}

	if fo.In().Map("b").Type() != core.TYPE_ANY {
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

func TestBuiltin_Eval__BoolArith(t *testing.T) {
	fo, err := createOpEval(core.InstanceDef{Operator: "eval", Properties: map[string]interface{}{"expression": "a && (b != c)"}}, nil)

	if fo == nil {
		t.Error("operator not defined")
	}

	if err != nil {
		t.Error(err.Error())
	}

	if fo.In().Type() != core.TYPE_MAP {
		t.Error("expected map")
	}

	if fo.In().Map("a").Type() != core.TYPE_ANY {
		t.Error("expected any input")
	}

	if fo.In().Map("b").Type() != core.TYPE_ANY {
		t.Error("expected any input")
	}

	if fo.In().Map("c").Type() != core.TYPE_ANY {
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
