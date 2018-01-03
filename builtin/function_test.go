package builtin

import (
	"testing"
	"slang"
)

func TestManager_MakeOperator__Function__NilProperties(t *testing.T) {
	fo, err := M().MakeOperator("function", nil)

	if fo != nil || err == nil {
		t.Error("expected error")
	}
}

func TestManager_MakeOperator__Function__Add(t *testing.T) {
	fo, err := M().MakeOperator("function", map[string]interface{}{"function": "a+b"})

	if fo == nil {
		t.Error("operator not defined")
	}

	if err != nil {
		t.Error(err.Error())
	}

	if fo.InPort().Type() != slang.TYPE_MAP {
		t.Error("expected map")
	}

	if fo.InPort().Port("a").Type() != slang.TYPE_ANY {
		t.Error("expected any input")
	}

	if fo.InPort().Port("b").Type() != slang.TYPE_ANY {
		t.Error("expected any input")
	}

	fo.OutPort().Bufferize()

	go fo.Start()

	fo.InPort().Push(map[string]interface{}{"a": 1.0, "b": 2.0})
	fo.InPort().Push(map[string]interface{}{"a": -5.0, "b": 2.5})
	fo.InPort().Push(map[string]interface{}{"a": 0.0, "b": 333.0})

	if fo.OutPort().Pull() != 3.0 {
		t.Error("wrong output")
	}

	if fo.OutPort().Pull() != -2.5 {
		t.Error("wrong output")
	}

	if fo.OutPort().Pull() != 333.0 {
		t.Error("wrong output")
	}
}

func TestManager_MakeOperator__Function__And(t *testing.T) {
	fo, err := M().MakeOperator("function", map[string]interface{}{"function": "a && (b || !c)"})

	if fo == nil {
		t.Error("operator not defined")
	}

	if err != nil {
		t.Error(err.Error())
	}

	if fo.InPort().Type() != slang.TYPE_MAP {
		t.Error("expected map")
	}

	if fo.InPort().Port("a").Type() != slang.TYPE_ANY {
		t.Error("expected any input")
	}

	if fo.InPort().Port("b").Type() != slang.TYPE_ANY {
		t.Error("expected any input")
	}

	if fo.InPort().Port("c").Type() != slang.TYPE_ANY {
		t.Error("expected any input")
	}

	fo.OutPort().Bufferize()

	go fo.Start()

	fo.InPort().Push(map[string]interface{}{"a": true, "b": true, "c": true})
	fo.InPort().Push(map[string]interface{}{"a": false, "b": true, "c": true})
	fo.InPort().Push(map[string]interface{}{"a": true, "b": false, "c": false})

	if fo.OutPort().Pull() != true {
		t.Error("wrong output")
	}

	if fo.OutPort().Pull() != false {
		t.Error("wrong output")
	}

	if fo.OutPort().Pull() != true {
		t.Error("wrong output")
	}
}
