package tests

import (
	"slang"
	"slang/op"
	"slang/tests/assertions"
	"testing"
)

// PortDef.Validate (11 tests)

func TestPortDef_Validate__InvalidTypeInDefinition(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"gfdhgfd"}`)
	p, err := op.NewPort(nil, def, op.DIRECTION_IN)
	a.Error(err)
	a.Nil(p)
}

func TestPortDef_Validate__Stream__StreamNotPresent(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"stream"}`)
	a.Error(def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Stream__NilStream(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"stream","stream":null}`)
	a.Error(def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Stream__EmptyStream(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"stream","stream":{}}`)
	a.Error(def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Stream__InvalidTypeInDefinition(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"stream","stream":{"type":"hgfdh"}}`)
	a.Error(def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Map__MapNotPresent(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"map"}`)
	a.Error(def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Map__NilMap(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"map","map":null}`)
	a.Error(def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Map__EmptyMap(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"map","map":{}}`)
	a.Error(def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Map__NullEntry(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"map","map":{"a":null}}`)
	a.Error(def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Map__EmptyEntry(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"map","map":{"a":{}}}`)
	a.Error(def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Map__InvalidTypeInDefinition(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"gfgfd"}}}`)
	a.Error(def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

// op.NewPort (10 tests)

func TestNewPort__InvalidDefinition(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"bcvbvcbvc"}`)
	p, err := op.NewPort(nil, def, 0)
	a.Error(err)
	a.Nil(p)
}

func TestNewPort__Number__NoDirectionGiven(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"number"}`)
	p, err := op.NewPort(nil, def, 0)
	a.Error(err)
	a.Nil(p)
}

func TestNewPort__Number__WrongDirectionGiven(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"number"}`)
	p, err := op.NewPort(nil, def, 3)
	a.Error(err)
	a.Nil(p)
}

func TestNewPort__Stream(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"stream","stream":{"type":"number"}}`)
	p, err := op.NewPort(nil, def, op.DIRECTION_IN)
	a.NoError(err)

	if p.Type() != op.TYPE_STREAM {
		t.Error("wrong type")
	}

	if p.Stream().Type() != op.TYPE_NUMBER {
		t.Error("wrong type")
	}
}

func TestNewPort__Number(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"number"}`)
	p, err := op.NewPort(nil, def, op.DIRECTION_IN)

	a.NoError(err)
	a.NotNil(p)

	if p.Type() != op.TYPE_NUMBER {
		t.Error("wrong type")
	}

	if p.Direction() != op.DIRECTION_IN {
		t.Error("Wrong direction")
	}
}

func TestNewPort__Map(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	p, err := op.NewPort(nil, def, op.DIRECTION_IN)
	a.NoError(err)

	if p.Type() != op.TYPE_MAP {
		t.Error("wrong type")
	}

	if p.Map("a").Type() != op.TYPE_NUMBER {
		t.Error("wrong type")
	}
}

func TestNewPort__NestedStreams(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"stream","stream":{"type":"stream","stream":{"type":"string"}}}`)
	p, err := op.NewPort(nil, def, op.DIRECTION_IN)
	a.NoError(err)

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

func TestNewPort__MapStream(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(
		`{"type":"map","map":{"a":{"type":"stream","stream":{"type":"string"}},"b":{"type":"boolean"}}}`)
	p, err := op.NewPort(nil, def, op.DIRECTION_IN)
	a.NoError(err)

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

func TestNewPort__NestedMap(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"},"b":{"type":"map","map":{"a":{"type":"number"}}}}}`)
	p, err := op.NewPort(nil, def, op.DIRECTION_IN)
	a.NoError(err)
	a.NotNil(p)

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

func TestNewPort__Complex(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(
		`{"type":"map","map":{"a":{"type":"stream","stream":{"type":"boolean"}},"b":{"type":"map","map":
{"a":{"type":"stream","stream":{"type":"stream","stream":{"type":"map","map":{"a":{"type":"number"},
"b":{"type":"string"},"c":{"type":"boolean"}}}}},"b":{"type":"string"}}},"c":{"type":"boolean"}}}`)
	p, err := op.NewPort(nil, def, op.DIRECTION_IN)
	a.NoError(err)

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
	def := slang.ParsePortDef(`{"type":"any"}`)
	p, _ := op.NewPort(nil, def, op.DIRECTION_IN)

	if p.Type() != op.TYPE_ANY {
		t.Error("wrong type")
	}
}

func TestPort_Type__Simple__Number(t *testing.T) {
	def := slang.ParsePortDef(`{"type":"number"}`)
	p, _ := op.NewPort(nil, def, op.DIRECTION_IN)

	if p.Type() != op.TYPE_NUMBER {
		t.Error("wrong type")
	}
}

func TestPort_Type__Simple__String(t *testing.T) {
	def := slang.ParsePortDef(`{"type":"string"}`)
	p, _ := op.NewPort(nil, def, op.DIRECTION_IN)

	if p.Type() != op.TYPE_STRING {
		t.Error("wrong type")
	}
}

func TestPort_Type__Simple__Boolean(t *testing.T) {
	def := slang.ParsePortDef(`{"type":"boolean"}`)
	p, _ := op.NewPort(nil, def, op.DIRECTION_IN)

	if p.Type() != op.TYPE_BOOLEAN {
		t.Error("wrong type")
	}
}

func TestPort_Type__Map(t *testing.T) {
	def := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"boolean"}}}`)
	p, _ := op.NewPort(nil, def, op.DIRECTION_IN)

	if p.Type() != op.TYPE_MAP {
		t.Error("wrong type")
	}

	if p.Map("a").Type() != op.TYPE_BOOLEAN {
		t.Error("wrong type")
	}
}

func TestPort_Type__Stream(t *testing.T) {
	def := slang.ParsePortDef(`{"type":"stream","stream":{"type":"string"}}`)
	p, _ := op.NewPort(nil, def, op.DIRECTION_IN)

	if p.Type() != op.TYPE_STREAM {
		t.Error("wrong type")
	}

	if p.Stream().Type() != op.TYPE_STRING {
		t.Error("wrong type")
	}
}

// Port.Connect (6 tests)

func TestPort_Connect__Map__KeysNotMatching(t *testing.T) {
	a := assertions.New(t)
	def1 := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := slang.ParsePortDef(`{"type":"map","map":{"b":{"type":"number"}}}`)

	p, _ := op.NewPort(nil, def1, op.DIRECTION_IN)
	q, _ := op.NewPort(nil, def2, op.DIRECTION_OUT)

	err := p.Connect(q)
	a.Error(err)

	if p.Connected(q) {
		t.Error("connection not expected")
	}
}

func TestPort_Connect__Map__LeftIsSubsetOfRight(t *testing.T) {
	a := assertions.New(t)
	def1 := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"},"b":{"type":"number"}}}`)

	p, _ := op.NewPort(nil, def1, op.DIRECTION_IN)
	q, _ := op.NewPort(nil, def2, op.DIRECTION_OUT)

	err := p.Connect(q)
	a.Error(err)

	if p.Connected(q) {
		t.Error("connection not expected")
	}
}

func TestPort_Connect__Map__RightIsSubsetOfRight(t *testing.T) {
	a := assertions.New(t)
	def1 := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"},"b":{"type":"number"}}}`)
	def2 := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)

	p, _ := op.NewPort(nil, def1, op.DIRECTION_IN)
	q, _ := op.NewPort(nil, def2, op.DIRECTION_OUT)

	err := p.Connect(q)
	a.Error(err)

	if p.Connected(q) {
		t.Error("connection not expected")
	}
}

func TestPort_Connect__Map__IncompatibleTypes(t *testing.T) {
	a := assertions.New(t)
	def1 := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := slang.ParsePortDef(`{"type":"number"}`)

	p, _ := op.NewPort(nil, def1, op.DIRECTION_IN)
	q, _ := op.NewPort(nil, def2, op.DIRECTION_OUT)

	err := p.Connect(q)
	a.Error(err)

	if p.Connected(q) {
		t.Error("connection not expected")
	}
}

func TestPort_Connect__Map__SameKeys(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)

	p, _ := op.NewPort(nil, def, op.DIRECTION_IN)
	q, _ := op.NewPort(nil, def, op.DIRECTION_OUT)

	err := p.Connect(q)
	a.NoError(err)

	if !p.Map("a").Connected(q.Map("a")) {
		t.Error("port 'a' must be connected")
	}

	if p.Connected(q) {
		t.Error("maps must not be connected")
	}
}

func TestPort_Connect__Map__Subport2Primitive(t *testing.T) {
	a := assertions.New(t)
	def1 := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := slang.ParsePortDef(`{"type":"number"}`)

	p, _ := op.NewPort(nil, def1, op.DIRECTION_IN)
	q, _ := op.NewPort(nil, def2, op.DIRECTION_OUT)

	err := p.Map("a").Connect(q)
	a.NoError(err)

	if !p.Map("a").Connected(q) {
		t.Error("connection expected")
	}
}
