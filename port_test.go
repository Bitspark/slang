package main

import "testing"

func TestMakePort__Map(t *testing.T) {
	def := helperJson2Map(`{"type":"map","map":{"a":{"type":"number"}}}`)
	p, err := MakePort(nil, def, DIRECTION_IN)
	assertNoError(t, err)
	assertNotNil(t, p)

	if p.Type() != TYPE_MAP {
		t.Error("type not correct")
	}

	if p.Port("a") == nil || p.Port("a").Type() != TYPE_ANY {
		t.Error("map not correct")
	}
}

func TestMakePort__NestedMap(t *testing.T) {
	def := helperJson2Map(`{"type":"map","map":{"a":{"type":"number"},"b":{"type":"map","map":{"a":{"type":"number"}}}}}`)
	p, err := MakePort(nil, def, DIRECTION_IN)
	assertNoError(t, err)
	assertNotNil(t, p)

	if p.Port("a") == nil || p.Port("a").Type() != TYPE_ANY {
		t.Error("map not correct")
	}

	if p.Port("b") == nil || p.Port("b").Type() != TYPE_MAP {
		t.Error("map not correct")
	}

	if p.Port("b").Port("a") == nil || p.Port("b").Port("a").Type() != TYPE_ANY {
		t.Error("map not correct")
	}
}

func TestPort_Connect__Map__KeysNotMatching(t *testing.T) {
	def1 := helperJson2Map(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := helperJson2Map(`{"type":"map","map":{"b":{"type":"number"}}}`)

	p, _ := MakePort(nil, def1, DIRECTION_IN)
	q, _ := MakePort(nil, def2, DIRECTION_OUT)

	err := p.Connect(q)
	assertError(t, err)
}

func TestPort_Connect__Map__IncompatibleTypes(t *testing.T) {
	def1 := helperJson2Map(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := helperJson2Map(`{"type":"number"}`)

	p, _ := MakePort(nil, def1, DIRECTION_IN)
	q, _ := MakePort(nil, def2, DIRECTION_OUT)

	err := p.Connect(q)
	assertError(t, err)
}

func TestPort_Connect__Map__SameKeys(t *testing.T) {
	def := helperJson2Map(`{"type":"map","map":{"a":{"type":"number"}}}`)

	p, _ := MakePort(nil, def, DIRECTION_IN)
	q, _ := MakePort(nil, def, DIRECTION_OUT)

	err := p.Connect(q)
	assertNoError(t, err)
}

func TestPort_Connect__Map__Subport2Primitive(t *testing.T) {
	def1 := helperJson2Map(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := helperJson2Map(`{"type":"number"}`)

	p, _ := MakePort(nil, def1, DIRECTION_IN)
	q, _ := MakePort(nil, def2, DIRECTION_OUT)

	err := p.Port("a").Connect(q)
	assertError(t, err)
}
