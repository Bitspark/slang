package tests

import (
	"slang"
	"slang/op"
	"testing"

	"github.com/stretchr/testify/assert"
)

// PortDef.Validate (11 tests)

func TestPortDef_Validate__InvalidTypeInDefinition(t *testing.T) {
	def := slang.ParsePortDef(`{"type":"gfdhgfd"}`)
	p, err := op.NewPort(nil, def, op.DIRECTION_IN)
	assert.Error(t, err)
	assert.Nil(t, p)
}

func TestPortDef_Validate__Stream__StreamNotPresent(t *testing.T) {
	def := slang.ParsePortDef(`{"type":"stream"}`)
	assert.Error(t, def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Stream__NilStream(t *testing.T) {
	def := slang.ParsePortDef(`{"type":"stream","stream":null}`)
	assert.Error(t, def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Stream__EmptyStream(t *testing.T) {
	def := slang.ParsePortDef(`{"type":"stream","stream":{}}`)
	assert.Error(t, def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Stream__InvalidTypeInDefinition(t *testing.T) {
	def := slang.ParsePortDef(`{"type":"stream","stream":{"type":"hgfdh"}}`)
	assert.Error(t, def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Map__MapNotPresent(t *testing.T) {
	def := slang.ParsePortDef(`{"type":"map"}`)
	assert.Error(t, def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Map__NilMap(t *testing.T) {
	def := slang.ParsePortDef(`{"type":"map","map":null}`)
	assert.Error(t, def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Map__EmptyMap(t *testing.T) {
	def := slang.ParsePortDef(`{"type":"map","map":{}}`)
	assert.Error(t, def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Map__NullEntry(t *testing.T) {
	def := slang.ParsePortDef(`{"type":"map","map":{"a":null}}`)
	assert.Error(t, def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Map__EmptyEntry(t *testing.T) {
	def := slang.ParsePortDef(`{"type":"map","map":{"a":{}}}`)
	assert.Error(t, def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

func TestPortDef_Validate__Map__InvalidTypeInDefinition(t *testing.T) {
	def := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"gfgfd"}}}`)
	assert.Error(t, def.Validate())
	if def.Valid() {
		t.Error("should not be valid")
	}
}

// op.NewPort (10 tests)

func TestNewPort__InvalidDefinition(t *testing.T) {
	def := slang.ParsePortDef(`{"type":"bcvbvcbvc"}`)
	p, err := op.NewPort(nil, def, 0)
	assert.Error(t, err)
	assert.Nil(t, p)
}

func TestNewPort__Number__NoDirectionGiven(t *testing.T) {
	def := slang.ParsePortDef(`{"type":"number"}`)
	p, err := op.NewPort(nil, def, 0)
	assert.Error(t, err)
	assert.Nil(t, p)
}

func TestNewPort__Number__WrongDirectionGiven(t *testing.T) {
	def := slang.ParsePortDef(`{"type":"number"}`)
	p, err := op.NewPort(nil, def, 3)
	assert.Error(t, err)
	assert.Nil(t, p)
}

func TestNewPort__Stream(t *testing.T) {
	def := slang.ParsePortDef(`{"type":"stream","stream":{"type":"number"}}`)
	p, err := op.NewPort(nil, def, op.DIRECTION_IN)
	assert.NoError(t, err)

	if p.Type() != op.TYPE_STREAM {
		t.Error("wrong type")
	}

	if p.Stream().Type() != op.TYPE_NUMBER {
		t.Error("wrong type")
	}
}

func TestNewPort__Number(t *testing.T) {
	def := slang.ParsePortDef(`{"type":"number"}`)
	p, err := op.NewPort(nil, def, op.DIRECTION_IN)

	assert.NoError(t, err)
	assert.NotNil(t, p)

	if p.Type() != op.TYPE_NUMBER {
		t.Error("wrong type")
	}

	if p.Direction() != op.DIRECTION_IN {
		t.Error("Wrong direction")
	}
}

func TestNewPort__Map(t *testing.T) {
	def := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	p, err := op.NewPort(nil, def, op.DIRECTION_IN)
	assert.NoError(t, err)

	if p.Type() != op.TYPE_MAP {
		t.Error("wrong type")
	}

	if p.Map("a").Type() != op.TYPE_NUMBER {
		t.Error("wrong type")
	}
}

func TestNewPort__NestedStreams(t *testing.T) {
	def := slang.ParsePortDef(`{"type":"stream","stream":{"type":"stream","stream":{"type":"string"}}}`)
	p, err := op.NewPort(nil, def, op.DIRECTION_IN)
	assert.NoError(t, err)

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
	def := slang.ParsePortDef(
		`{"type":"map","map":{"a":{"type":"stream","stream":{"type":"string"}},"b":{"type":"boolean"}}}`)
	p, err := op.NewPort(nil, def, op.DIRECTION_IN)
	assert.NoError(t, err)

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
	def := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"},"b":{"type":"map","map":{"a":{"type":"number"}}}}}`)
	p, err := op.NewPort(nil, def, op.DIRECTION_IN)
	assert.NoError(t, err)
	assert.NotNil(t, p)

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
	def := slang.ParsePortDef(
		`{"type":"map","map":{"a":{"type":"stream","stream":{"type":"boolean"}},"b":{"type":"map","map":
{"a":{"type":"stream","stream":{"type":"stream","stream":{"type":"map","map":{"a":{"type":"number"},
"b":{"type":"string"},"c":{"type":"boolean"}}}}},"b":{"type":"string"}}},"c":{"type":"boolean"}}}`)
	p, err := op.NewPort(nil, def, op.DIRECTION_IN)
	assert.NoError(t, err)

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
	def1 := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := slang.ParsePortDef(`{"type":"map","map":{"b":{"type":"number"}}}`)

	p, _ := op.NewPort(nil, def1, op.DIRECTION_IN)
	q, _ := op.NewPort(nil, def2, op.DIRECTION_OUT)

	err := p.Connect(q)
	assert.Error(t, err)

	if p.Connected(q) {
		t.Error("connection not expected")
	}
}

func TestPort_Connect__Map__LeftIsSubsetOfRight(t *testing.T) {
	def1 := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"},"b":{"type":"number"}}}`)

	p, _ := op.NewPort(nil, def1, op.DIRECTION_IN)
	q, _ := op.NewPort(nil, def2, op.DIRECTION_OUT)

	err := p.Connect(q)
	assert.Error(t, err)

	if p.Connected(q) {
		t.Error("connection not expected")
	}
}

func TestPort_Connect__Map__RightIsSubsetOfRight(t *testing.T) {
	def1 := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"},"b":{"type":"number"}}}`)
	def2 := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)

	p, _ := op.NewPort(nil, def1, op.DIRECTION_IN)
	q, _ := op.NewPort(nil, def2, op.DIRECTION_OUT)

	err := p.Connect(q)
	assert.Error(t, err)

	if p.Connected(q) {
		t.Error("connection not expected")
	}
}

func TestPort_Connect__Map__IncompatibleTypes(t *testing.T) {
	def1 := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := slang.ParsePortDef(`{"type":"number"}`)

	p, _ := op.NewPort(nil, def1, op.DIRECTION_IN)
	q, _ := op.NewPort(nil, def2, op.DIRECTION_OUT)

	err := p.Connect(q)
	assert.Error(t, err)

	if p.Connected(q) {
		t.Error("connection not expected")
	}
}

func TestPort_Connect__Map__SameKeys(t *testing.T) {
	def := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)

	p, _ := op.NewPort(nil, def, op.DIRECTION_IN)
	q, _ := op.NewPort(nil, def, op.DIRECTION_OUT)

	err := p.Connect(q)
	assert.NoError(t, err)

	if !p.Map("a").Connected(q.Map("a")) {
		t.Error("port 'a' must be connected")
	}

	if p.Connected(q) {
		t.Error("maps must not be connected")
	}
}

func TestPort_Connect__Map__Subport2Primitive(t *testing.T) {
	def1 := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := slang.ParsePortDef(`{"type":"number"}`)

	p, _ := op.NewPort(nil, def1, op.DIRECTION_IN)
	q, _ := op.NewPort(nil, def2, op.DIRECTION_OUT)

	err := p.Map("a").Connect(q)
	assert.NoError(t, err)

	if !p.Map("a").Connected(q) {
		t.Error("connection expected")
	}
}
