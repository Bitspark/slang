package tests

import (
	"slang/op"
	"testing"
)

// PortDef.Validate (11 tests)

func TestPortDef_Validate__InvalidTypeInDefinition(t *testing.T) {
	def := op.ParsePortDef(`{"type":"gfdhgfd"}`)
	p, err := op.MakePort(nil, def, op.DIRECTION_IN)
	assertError(t, err)
	assertNil(t, p)
}

func TestPortDef_Validate__Stream__StreamNotPresent(t *testing.T) {
	def := op.ParsePortDef(`{"type":"stream"}`)
	assertError(t, def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Stream__NilStream(t *testing.T) {
	def := op.ParsePortDef(`{"type":"stream","stream":null}`)
	assertError(t, def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Stream__EmptyStream(t *testing.T) {
	def := op.ParsePortDef(`{"type":"stream","stream":{}}`)
	assertError(t, def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Stream__InvalidTypeInDefinition(t *testing.T) {
	def := op.ParsePortDef(`{"type":"stream","stream":{"type":"hgfdh"}}`)
	assertError(t, def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Map__MapNotPresent(t *testing.T) {
	def := op.ParsePortDef(`{"type":"map"}`)
	assertError(t, def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Map__NilMap(t *testing.T) {
	def := op.ParsePortDef(`{"type":"map","map":null}`)
	assertError(t, def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Map__EmptyMap(t *testing.T) {
	def := op.ParsePortDef(`{"type":"map","map":{}}`)
	assertError(t, def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Map__NullEntry(t *testing.T) {
	def := op.ParsePortDef(`{"type":"map","map":{"a":null}}`)
	assertError(t, def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Map__EmptyEntry(t *testing.T) {
	def := op.ParsePortDef(`{"type":"map","map":{"a":{}}}`)
	assertError(t, def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Map__InvalidTypeInDefinition(t *testing.T) {
	def := op.ParsePortDef(`{"type":"map","map":{"a":{"type":"gfgfd"}}}`)
	assertError(t, def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

// op.MakePort (10 tests)

func TestMakePort__InvalidDefinition(t *testing.T) {
	def := op.ParsePortDef(`{"type":"bcvbvcbvc"}`)
	p, err := op.MakePort(nil, def, 0)
	assertError(t, err)
	assertNil(t, p)
}

func TestMakePort__Number__NoDirectionGiven(t *testing.T) {
	def := op.ParsePortDef(`{"type":"number"}`)
	p, err := op.MakePort(nil, def, 0)
	assertError(t, err)
	assertNil(t, p)
}

func TestMakePort__Number__WrongDirectionGiven(t *testing.T) {
	def := op.ParsePortDef(`{"type":"number"}`)
	p, err := op.MakePort(nil, def, 3)
	assertError(t, err)
	assertNil(t, p)
}

func TestMakePort__Stream(t *testing.T) {
	def := op.ParsePortDef(`{"type":"stream","stream":{"type":"number"}}`)
	p, err := op.MakePort(nil, def, op.DIRECTION_IN)
	assertNoError(t, err)

	if p.Type() != op.TYPE_STREAM {
		t.Error("wrong type")
	}

	if p.Stream().Type() != op.TYPE_NUMBER {
		t.Error("wrong type")
	}
}

func TestMakePort__Number(t *testing.T) {
	def := op.ParsePortDef(`{"type":"number"}`)
	p, err := op.MakePort(nil, def, op.DIRECTION_IN)

	assertNoError(t, err)
	assertNotNil(t, p)

	if p.Type() != op.TYPE_NUMBER {
		t.Error("wrong type")
	}

	if p.Direction() != op.DIRECTION_IN {
		t.Error("Wrong direction")
	}
}

func TestMakePort__Map(t *testing.T) {
	def := op.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	p, err := op.MakePort(nil, def, op.DIRECTION_IN)
	assertNoError(t, err)

	if p.Type() != op.TYPE_MAP {
		t.Error("wrong type")
	}

	if p.Map("a").Type() != op.TYPE_NUMBER {
		t.Error("wrong type")
	}
}

func TestMakePort__NestedStreams(t *testing.T) {
	def := op.ParsePortDef(`{"type":"stream","stream":{"type":"stream","stream":{"type":"string"}}}`)
	p, err := op.MakePort(nil, def, op.DIRECTION_IN)
	assertNoError(t, err)

	if p.Type() != op.TYPE_STREAM {
		t.Error("wrong type")
	}

	if p.Stream().Type() != op.TYPE_STREAM {
		t.Error("wrong type")
	}

	if p.Stream().Stream().Type() != op.TYPE_STRING {
		t.Error("wrong type")
	}
}

func TestMakePort__MapStream(t *testing.T) {
	def := op.ParsePortDef(
		`{"type":"map","map":{"a":{"type":"stream","stream":{"type":"string"}},"b":{"type":"boolean"}}}`)
	p, err := op.MakePort(nil, def, op.DIRECTION_IN)
	assertNoError(t, err)

	if p.Type() != op.TYPE_MAP {
		t.Error("wrong type")
	}

	if p.Map("a").Type() != op.TYPE_STREAM {
		t.Error("wrong type")
	}

	if p.Map("a").Stream().Type() != op.TYPE_STRING {
		t.Error("wrong type")
	}

	if p.Map("b").Type() != op.TYPE_BOOLEAN {
		t.Error("wrong type")
	}
}

func TestMakePort__NestedMap(t *testing.T) {
	def := op.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"},"b":{"type":"map","map":{"a":{"type":"number"}}}}}`)
	p, err := op.MakePort(nil, def, op.DIRECTION_IN)
	assertNoError(t, err)
	assertNotNil(t, p)

	if p.Map("a") == nil || p.Map("a").Type() != op.TYPE_NUMBER {
		t.Error("map not correct")
	}

	if p.Map("b") == nil || p.Map("b").Type() != op.TYPE_MAP {
		t.Error("map not correct")
	}

	if p.Map("b").Map("a") == nil || p.Map("b").Map("a").Type() != op.TYPE_NUMBER {
		t.Error("map not correct")
	}
}

func TestMakePort__Complex(t *testing.T) {
	def := op.ParsePortDef(
		`{"type":"map","map":{"a":{"type":"stream","stream":{"type":"boolean"}},"b":{"type":"map","map":
{"a":{"type":"stream","stream":{"type":"stream","stream":{"type":"map","map":{"a":{"type":"number"},
"b":{"type":"string"},"c":{"type":"boolean"}}}}},"b":{"type":"string"}}},"c":{"type":"boolean"}}}`)
	p, err := op.MakePort(nil, def, op.DIRECTION_IN)
	assertNoError(t, err)

	if p.Type() != op.TYPE_MAP {
		t.Error("wrong type")
	}

	if p.Map("a").Type() != op.TYPE_STREAM {
		t.Error("wrong type")
	}

	if p.Map("a").Stream().Type() != op.TYPE_BOOLEAN {
		t.Error("wrong type")
	}

	if p.Map("b").Type() != op.TYPE_MAP {
		t.Error("wrong type")
	}

	if p.Map("b").Map("a").Type() != op.TYPE_STREAM {
		t.Error("wrong type")
	}

	if p.Map("b").Map("a").Stream().Type() != op.TYPE_STREAM {
		t.Error("wrong type")
	}

	if p.Map("b").Map("a").Stream().Stream().Type() != op.TYPE_MAP {
		t.Error("wrong type")
	}

	if p.Map("b").Map("a").Stream().Stream().Map("a").Type() != op.TYPE_NUMBER {
		t.Error("wrong type")
	}

	if p.Map("b").Map("a").Stream().Stream().Map("b").Type() != op.TYPE_STRING {
		t.Error("wrong type")
	}

	if p.Map("b").Map("a").Stream().Stream().Map("c").Type() != op.TYPE_BOOLEAN {
		t.Error("wrong type")
	}

	if p.Map("b").Map("b").Type() != op.TYPE_STRING {
		t.Error("wrong type")
	}

	if p.Map("c").Type() != op.TYPE_BOOLEAN {
		t.Error("wrong type")
	}
}

// Port.Type (6 tests)

func TestPort_Type__Simple__Any(t *testing.T) {
	def := op.ParsePortDef(`{"type":"any"}`)
	p, _ := op.MakePort(nil, def, op.DIRECTION_IN)

	if p.Type() != op.TYPE_ANY {
		t.Error("wrong type")
	}
}

func TestPort_Type__Simple__Number(t *testing.T) {
	def := op.ParsePortDef(`{"type":"number"}`)
	p, _ := op.MakePort(nil, def, op.DIRECTION_IN)

	if p.Type() != op.TYPE_NUMBER {
		t.Error("wrong type")
	}
}

func TestPort_Type__Simple__String(t *testing.T) {
	def := op.ParsePortDef(`{"type":"string"}`)
	p, _ := op.MakePort(nil, def, op.DIRECTION_IN)

	if p.Type() != op.TYPE_STRING {
		t.Error("wrong type")
	}
}

func TestPort_Type__Simple__Boolean(t *testing.T) {
	def := op.ParsePortDef(`{"type":"boolean"}`)
	p, _ := op.MakePort(nil, def, op.DIRECTION_IN)

	if p.Type() != op.TYPE_BOOLEAN {
		t.Error("wrong type")
	}
}

func TestPort_Type__Map(t *testing.T) {
	def := op.ParsePortDef(`{"type":"map","map":{"a":{"type":"boolean"}}}`)
	p, _ := op.MakePort(nil, def, op.DIRECTION_IN)

	if p.Type() != op.TYPE_MAP {
		t.Error("wrong type")
	}

	if p.Map("a").Type() != op.TYPE_BOOLEAN {
		t.Error("wrong type")
	}
}

func TestPort_Type__Stream(t *testing.T) {
	def := op.ParsePortDef(`{"type":"stream","stream":{"type":"string"}}`)
	p, _ := op.MakePort(nil, def, op.DIRECTION_IN)

	if p.Type() != op.TYPE_STREAM {
		t.Error("wrong type")
	}

	if p.Stream().Type() != op.TYPE_STRING {
		t.Error("wrong type")
	}
}

// Port.Connect (6 tests)

func TestPort_Connect__Map__KeysNotMatching(t *testing.T) {
	def1 := op.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := op.ParsePortDef(`{"type":"map","map":{"b":{"type":"number"}}}`)

	p, _ := op.MakePort(nil, def1, op.DIRECTION_IN)
	q, _ := op.MakePort(nil, def2, op.DIRECTION_OUT)

	err := p.Connect(q)
	assertError(t, err)

	if p.Connected(q) {
		t.Error("connection not expected")
	}
}

func TestPort_Connect__Map__LeftIsSubsetOfRight(t *testing.T) {
	def1 := op.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := op.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"},"b":{"type":"number"}}}`)

	p, _ := op.MakePort(nil, def1, op.DIRECTION_IN)
	q, _ := op.MakePort(nil, def2, op.DIRECTION_OUT)

	err := p.Connect(q)
	assertError(t, err)

	if p.Connected(q) {
		t.Error("connection not expected")
	}
}

func TestPort_Connect__Map__RightIsSubsetOfRight(t *testing.T) {
	def1 := op.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"},"b":{"type":"number"}}}`)
	def2 := op.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)

	p, _ := op.MakePort(nil, def1, op.DIRECTION_IN)
	q, _ := op.MakePort(nil, def2, op.DIRECTION_OUT)

	err := p.Connect(q)
	assertError(t, err)

	if p.Connected(q) {
		t.Error("connection not expected")
	}
}

func TestPort_Connect__Map__IncompatibleTypes(t *testing.T) {
	def1 := op.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := op.ParsePortDef(`{"type":"number"}`)

	p, _ := op.MakePort(nil, def1, op.DIRECTION_IN)
	q, _ := op.MakePort(nil, def2, op.DIRECTION_OUT)

	err := p.Connect(q)
	assertError(t, err)

	if p.Connected(q) {
		t.Error("connection not expected")
	}
}

func TestPort_Connect__Map__SameKeys(t *testing.T) {
	def := op.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)

	p, _ := op.MakePort(nil, def, op.DIRECTION_IN)
	q, _ := op.MakePort(nil, def, op.DIRECTION_OUT)

	err := p.Connect(q)
	assertNoError(t, err)

	if !p.Map("a").Connected(q.Map("a")) {
		t.Error("port 'a' must be connected")
	}

	if p.Connected(q) {
		t.Error("maps must not be connected")
	}
}

func TestPort_Connect__Map__Subport2Primitive(t *testing.T) {
	def1 := op.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := op.ParsePortDef(`{"type":"number"}`)

	p, _ := op.MakePort(nil, def1, op.DIRECTION_IN)
	q, _ := op.MakePort(nil, def2, op.DIRECTION_OUT)

	err := p.Map("a").Connect(q)
	assertNoError(t, err)

	if !p.Map("a").Connected(q) {
		t.Error("connection expected")
	}
}