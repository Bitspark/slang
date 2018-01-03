package slang

import (
	"testing"
)

// MakePort (20 tests)

func TestMakePort__InvalidTypeInDefinition(t *testing.T) {
	def := helperJson2PortDef(`{"type":"gfdhgfd"}`)
	p, err := MakePort(nil, def, DIRECTION_IN)
	assertError(t, err)
	assertNil(t, p)
}

func TestMakePort__Number__NoDirectionGiven(t *testing.T) {
	def := helperJson2PortDef(`{"type":"number"}`)
	p, err := MakePort(nil, def, 0)
	assertError(t, err)
	assertNil(t, p)
}

func TestMakePort__Number__WrongDirectionGiven(t *testing.T) {
	def := helperJson2PortDef(`{"type":"number"}`)
	p, err := MakePort(nil, def, 3)
	assertError(t, err)
	assertNil(t, p)
}

func TestMakePort__Stream__StreamNotPresent(t *testing.T) {
	def := helperJson2PortDef(`{"type":"stream"}`)
	p, err := MakePort(nil, def, DIRECTION_IN)
	assertError(t, err)
	assertNil(t, p)
}

func TestMakePort__Stream__NilStream(t *testing.T) {
	def := helperJson2PortDef(`{"type":"stream","stream":null}`)
	p, err := MakePort(nil, def, DIRECTION_IN)
	assertError(t, err)
	assertNil(t, p)
}

func TestMakePort__Stream__EmptyStream(t *testing.T) {
	def := helperJson2PortDef(`{"type":"stream","stream":{}}`)
	p, err := MakePort(nil, def, DIRECTION_IN)
	assertError(t, err)
	assertNil(t, p)
}

func TestMakePort__Stream__InvalidTypeInDefinition(t *testing.T) {
	def := helperJson2PortDef(`{"type":"stream","stream":{"type":"hgfdh"}}`)
	p, err := MakePort(nil, def, DIRECTION_IN)
	assertError(t, err)
	assertNil(t, p)
}

func TestMakePort__Stream__Success(t *testing.T) {
	def := helperJson2PortDef(`{"type":"stream","stream":{"type":"number"}}`)
	p, err := MakePort(nil, def, DIRECTION_IN)
	assertNoError(t, err)

	if p.Type() != TYPE_STREAM {
		t.Error("wrong type")
	}

	if p.Stream().Type() != TYPE_NUMBER {
		t.Error("wrong type")
	}
}

func TestMakePort__Map__MapNotPresent(t *testing.T) {
	def := helperJson2PortDef(`{"type":"map"}`)
	p, err := MakePort(nil, def, DIRECTION_IN)
	assertError(t, err)
	assertNil(t, p)
}

func TestMakePort__Map__NilMap(t *testing.T) {
	def := helperJson2PortDef(`{"type":"map","map":null}`)
	p, err := MakePort(nil, def, DIRECTION_IN)
	assertError(t, err)
	assertNil(t, p)
}

func TestMakePort__Map__EmptyMap(t *testing.T) {
	def := helperJson2PortDef(`{"type":"map","map":{}}`)
	p, err := MakePort(nil, def, DIRECTION_IN)
	assertError(t, err)
	assertNil(t, p)
}

func TestMakePort__Map__NullEntry(t *testing.T) {
	def := helperJson2PortDef(`{"type":"map","map":{"a":null}}`)
	p, err := MakePort(nil, def, DIRECTION_IN)
	assertError(t, err)
	assertNil(t, p)
}

func TestMakePort__Map__EmptyEntry(t *testing.T) {
	def := helperJson2PortDef(`{"type":"map","map":{"a":{}}}`)
	p, err := MakePort(nil, def, DIRECTION_IN)
	assertError(t, err)
	assertNil(t, p)
}

func TestMakePort__Map__InvalidTypeInDefinition(t *testing.T) {
	def := helperJson2PortDef(`{"type":"map","map":{"a":{"type":"gfgfd"}}}`)
	p, err := MakePort(nil, def, DIRECTION_IN)
	assertError(t, err)
	assertNil(t, p)
}

func TestMakePort__Number__Success(t *testing.T) {
	def := helperJson2PortDef(`{"type":"number"}`)
	p, err := MakePort(nil, def, DIRECTION_IN)

	assertNoError(t, err)
	assertNotNil(t, p)

	if p.Type() != TYPE_NUMBER {
		t.Error("wrong type")
	}

	if p.Direction() != DIRECTION_IN {
		t.Error("Wrong direction")
	}
}

func TestMakePort__Map__Success(t *testing.T) {
	def := helperJson2PortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	p, err := MakePort(nil, def, DIRECTION_IN)
	assertNoError(t, err)

	if p.Type() != TYPE_MAP {
		t.Error("wrong type")
	}

	if p.Port("a").Type() != TYPE_NUMBER {
		t.Error("wrong type")
	}
}

func TestMakePort__NestedStreams__Success(t *testing.T) {
	def := helperJson2PortDef(`{"type":"stream","stream":{"type":"stream","stream":{"type":"string"}}}`)
	p, err := MakePort(nil, def, DIRECTION_IN)
	assertNoError(t, err)

	if p.Type() != TYPE_STREAM {
		t.Error("wrong type")
	}

	if p.Stream().Type() != TYPE_STREAM {
		t.Error("wrong type")
	}

	if p.Stream().Stream().Type() != TYPE_STRING {
		t.Error("wrong type")
	}
}

func TestMakePort__MapStream__Success(t *testing.T) {
	def := helperJson2PortDef(
`{"type":"map","map":{"a":{"type":"stream","stream":{"type":"string"}},"b":{"type":"boolean"}}}`)
	p, err := MakePort(nil, def, DIRECTION_IN)
	assertNoError(t, err)

	if p.Type() != TYPE_MAP {
		t.Error("wrong type")
	}

	if p.Port("a").Type() != TYPE_STREAM {
		t.Error("wrong type")
	}

	if p.Port("a").Stream().Type() != TYPE_STRING {
		t.Error("wrong type")
	}

	if p.Port("b").Type() != TYPE_BOOLEAN {
		t.Error("wrong type")
	}
}

func TestMakePort__NestedMap__Success(t *testing.T) {
	def := helperJson2PortDef(`{"type":"map","map":{"a":{"type":"number"},"b":{"type":"map","map":{"a":{"type":"number"}}}}}`)
	p, err := MakePort(nil, def, DIRECTION_IN)
	assertNoError(t, err)
	assertNotNil(t, p)

	if p.Port("a") == nil || p.Port("a").Type() != TYPE_NUMBER {
		t.Error("map not correct")
	}

	if p.Port("b") == nil || p.Port("b").Type() != TYPE_MAP {
		t.Error("map not correct")
	}

	if p.Port("b").Port("a") == nil || p.Port("b").Port("a").Type() != TYPE_NUMBER {
		t.Error("map not correct")
	}
}

func TestMakePort__Complex__Success(t *testing.T) {
	def := helperJson2PortDef(
`{"type":"map","map":{"a":{"type":"stream","stream":{"type":"boolean"}},"b":{"type":"map","map":
{"a":{"type":"stream","stream":{"type":"stream","stream":{"type":"map","map":{"a":{"type":"number"},
"b":{"type":"string"},"c":{"type":"boolean"}}}}},"b":{"type":"string"}}},"c":{"type":"boolean"}}}`)
	p, err := MakePort(nil, def, DIRECTION_IN)
	assertNoError(t, err)

	if p.Type() != TYPE_MAP {
		t.Error("wrong type")
	}

	if p.Port("a").Type() != TYPE_STREAM {
		t.Error("wrong type")
	}

	if p.Port("a").Stream().Type() != TYPE_BOOLEAN {
		t.Error("wrong type")
	}

	if p.Port("b").Type() != TYPE_MAP {
		t.Error("wrong type")
	}

	if p.Port("b").Port("a").Type() != TYPE_STREAM {
		t.Error("wrong type")
	}

	if p.Port("b").Port("a").Stream().Type() != TYPE_STREAM {
		t.Error("wrong type")
	}

	if p.Port("b").Port("a").Stream().Stream().Type() != TYPE_MAP {
		t.Error("wrong type")
	}

	if p.Port("b").Port("a").Stream().Stream().Port("a").Type() != TYPE_NUMBER {
		t.Error("wrong type")
	}

	if p.Port("b").Port("a").Stream().Stream().Port("b").Type() != TYPE_STRING {
		t.Error("wrong type")
	}

	if p.Port("b").Port("a").Stream().Stream().Port("c").Type() != TYPE_BOOLEAN {
		t.Error("wrong type")
	}

	if p.Port("b").Port("b").Type() != TYPE_STRING {
		t.Error("wrong type")
	}

	if p.Port("c").Type() != TYPE_BOOLEAN {
		t.Error("wrong type")
	}
}

// Port.Type (6 tests)

func TestPort_Type__Simple__Any(t *testing.T) {
	def := helperJson2PortDef(`{"type":"any"}`)
	p, _ := MakePort(nil, def, DIRECTION_IN)

	if p.Type() != TYPE_ANY {
		t.Error("wrong type")
	}
}

func TestPort_Type__Simple__Number(t *testing.T) {
	def := helperJson2PortDef(`{"type":"number"}`)
	p, _ := MakePort(nil, def, DIRECTION_IN)

	if p.Type() != TYPE_NUMBER {
		t.Error("wrong type")
	}
}

func TestPort_Type__Simple__String(t *testing.T) {
	def := helperJson2PortDef(`{"type":"string"}`)
	p, _ := MakePort(nil, def, DIRECTION_IN)

	if p.Type() != TYPE_STRING {
		t.Error("wrong type")
	}
}

func TestPort_Type__Simple__Boolean(t *testing.T) {
	def := helperJson2PortDef(`{"type":"boolean"}`)
	p, _ := MakePort(nil, def, DIRECTION_IN)

	if p.Type() != TYPE_BOOLEAN {
		t.Error("wrong type")
	}
}

func TestPort_Type__Map(t *testing.T) {
	def := helperJson2PortDef(`{"type":"map","map":{"a":{"type":"boolean"}}}`)
	p, _ := MakePort(nil, def, DIRECTION_IN)

	if p.Type() != TYPE_MAP {
		t.Error("wrong type")
	}

	if p.Port("a").Type() != TYPE_BOOLEAN {
		t.Error("wrong type")
	}
}

func TestPort_Type__Stream(t *testing.T) {
	def := helperJson2PortDef(`{"type":"stream","stream":{"type":"string"}}`)
	p, _ := MakePort(nil, def, DIRECTION_IN)

	if p.Type() != TYPE_STREAM {
		t.Error("wrong type")
	}

	if p.Stream().Type() != TYPE_STRING {
		t.Error("wrong type")
	}
}

// Port.Connect (6 tests)

func TestPort_Connect__Map__KeysNotMatching(t *testing.T) {
	def1 := helperJson2PortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := helperJson2PortDef(`{"type":"map","map":{"b":{"type":"number"}}}`)

	p, _ := MakePort(nil, def1, DIRECTION_IN)
	q, _ := MakePort(nil, def2, DIRECTION_OUT)

	err := p.Connect(q)
	assertError(t, err)

	if p.Connected(q) {
		t.Error("connection not expected")
	}
}

func TestPort_Connect__Map__LeftIsSubsetOfRight(t *testing.T) {
	def1 := helperJson2PortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := helperJson2PortDef(`{"type":"map","map":{"a":{"type":"number"},"b":{"type":"number"}}}`)

	p, _ := MakePort(nil, def1, DIRECTION_IN)
	q, _ := MakePort(nil, def2, DIRECTION_OUT)

	err := p.Connect(q)
	assertError(t, err)

	if p.Connected(q) {
		t.Error("connection not expected")
	}
}

func TestPort_Connect__Map__RightIsSubsetOfRight(t *testing.T) {
	def1 := helperJson2PortDef(`{"type":"map","map":{"a":{"type":"number"},"b":{"type":"number"}}}`)
	def2 := helperJson2PortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)

	p, _ := MakePort(nil, def1, DIRECTION_IN)
	q, _ := MakePort(nil, def2, DIRECTION_OUT)

	err := p.Connect(q)
	assertError(t, err)

	if p.Connected(q) {
		t.Error("connection not expected")
	}
}

func TestPort_Connect__Map__IncompatibleTypes(t *testing.T) {
	def1 := helperJson2PortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := helperJson2PortDef(`{"type":"number"}`)

	p, _ := MakePort(nil, def1, DIRECTION_IN)
	q, _ := MakePort(nil, def2, DIRECTION_OUT)

	err := p.Connect(q)
	assertError(t, err)

	if p.Connected(q) {
		t.Error("connection not expected")
	}
}

func TestPort_Connect__Map__SameKeys(t *testing.T) {
	def := helperJson2PortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)

	p, _ := MakePort(nil, def, DIRECTION_IN)
	q, _ := MakePort(nil, def, DIRECTION_OUT)

	err := p.Connect(q)
	assertNoError(t, err)

	if !p.Port("a").Connected(q.Port("a")) {
		t.Error("port 'a' must be connected")
	}

	if p.Connected(q) {
		t.Error("maps must not be connected")
	}
}

func TestPort_Connect__Map__Subport2Primitive(t *testing.T) {
	def1 := helperJson2PortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := helperJson2PortDef(`{"type":"number"}`)

	p, _ := MakePort(nil, def1, DIRECTION_IN)
	q, _ := MakePort(nil, def2, DIRECTION_OUT)

	err := p.Port("a").Connect(q)
	assertNoError(t, err)

	if !p.Port("a").Connected(q) {
		t.Error("connection expected")
	}
}
