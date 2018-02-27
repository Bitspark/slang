package tests

import (
	"slang"
	"slang/core"
	"slang/tests/assertions"
	"testing"
)

// PortDef.Validate (11 tests)

func TestPortDef_Validate__InvalidTypeInDefinition(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"gfdhgfd"}`)
	p, err := core.NewPort(nil, nil, def, core.DIRECTION_IN, nil)
	a.Error(err)
	a.Nil(p)
}

func TestPortDef_Validate__Stream__StreamNotPresent(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"stream"}`)
	a.Error(def.Validate())
	a.False(def.Valid(), "should not be valid")
}

func TestPortDef_Validate__Stream__NilStream(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"stream","stream":null}`)
	a.Error(def.Validate())
	a.False(def.Valid(), "should not be valid")
}

func TestPortDef_Validate__Stream__EmptyStream(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"stream","stream":{}}`)
	a.Error(def.Validate())
	a.False(def.Valid(), "should not be valid")
}

func TestPortDef_Validate__Stream__InvalidTypeInDefinition(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"stream","stream":{"type":"hgfdh"}}`)
	a.Error(def.Validate())
	a.False(def.Valid(), "should not be valid")
}

func TestPortDef_Validate__Map__MapNotPresent(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"map"}`)
	a.Error(def.Validate())
	a.False(def.Valid(), "should not be valid")
}

func TestPortDef_Validate__Map__NilMap(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"map","map":null}`)
	a.Error(def.Validate())
	a.False(def.Valid(), "should not be valid")
}

func TestPortDef_Validate__Map__EmptyMap(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"map","map":{}}`)
	a.Error(def.Validate())
	a.False(def.Valid(), "should not be valid")
}

func TestPortDef_Validate__Map__NullEntry(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"map","map":{"a":null}}`)
	a.Error(def.Validate())
	a.False(def.Valid(), "should not be valid")
}

func TestPortDef_Validate__Map__EmptyEntry(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"map","map":{"a":{}}}`)
	a.Error(def.Validate())
	a.False(def.Valid(), "should not be valid")
}

func TestPortDef_Validate__Map__InvalidTypeInDefinition(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"gfgfd"}}}`)
	a.Error(def.Validate())
	a.False(def.Valid(), "should not be valid")
}

func TestPortDef_Validate__Generic__IdentifierMissing(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"generic"}`)
	a.Error(def.Validate())
	a.False(def.Valid(), "should not be valid")
}

func TestPortDef_Validate__Generic__Correct(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"generic", "generic":"id1"}`)
	a.NoError(def.Validate())
	a.True(def.Valid(), "should be valid")
}

// core.NewPort (10 tests)

func TestNewPort__InvalidDefinition(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"bcvbvcbvc"}`)
	p, err := core.NewPort(nil, nil, def, 0, nil)
	a.Error(err)
	a.Nil(p)
}

func TestNewPort__Number__NoDirectionGiven(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"number"}`)
	p, err := core.NewPort(nil, nil, def, 0, nil)
	a.Error(err)
	a.Nil(p)
}

func TestNewPort__Number__WrongDirectionGiven(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"number"}`)
	p, err := core.NewPort(nil, nil, def, 3, nil)
	a.Error(err)
	a.Nil(p)
}

func TestNewPort__Stream(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"stream","stream":{"type":"number"}}`)
	p, err := core.NewPort(nil, nil, def, core.DIRECTION_IN, nil)
	a.NoError(err)
	a.Equal(core.TYPE_STREAM, p.Type(), "wrong type")
	a.Equal(core.TYPE_NUMBER, p.Stream().Type(), "wrong type")
}

func TestNewPort__Number(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"number"}`)
	p, err := core.NewPort(nil, nil, def, core.DIRECTION_IN, nil)

	a.NoError(err)
	a.NotNil(p)
	a.Equal(core.TYPE_NUMBER, p.Type(), "wrong type")
	a.Equal(core.DIRECTION_IN, p.Direction(), "wrong direction")
}

func TestNewPort__Map(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	p, err := core.NewPort(nil, nil, def, core.DIRECTION_IN, nil)
	a.NoError(err)
	a.Equal(core.TYPE_MAP, p.Type(), "wrong type")
	a.Equal(core.TYPE_NUMBER, p.Map("a").Type(), "wrong type")
}

func TestNewPort__NestedStreams(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"stream","stream":{"type":"stream","stream":{"type":"string"}}}`)
	p, err := core.NewPort(nil, nil, def, core.DIRECTION_IN, nil)
	a.NoError(err)
	a.Equal(core.TYPE_STREAM, p.Type(), "wrong type")
	a.Equal(core.TYPE_STREAM, p.Stream().Type(), "wrong type")
	a.Equal(core.TYPE_STRING, p.Stream().Stream().Type(), "wrong type")
}

func TestNewPort__MapStream(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(
		`{"type":"map","map":{"a":{"type":"stream","stream":{"type":"string"}},"b":{"type":"boolean"}}}`)
	p, err := core.NewPort(nil, nil, def, core.DIRECTION_IN, nil)
	a.NoError(err)
	a.Equal(core.TYPE_MAP, p.Type(), "wrong type")
	a.Equal(core.TYPE_STREAM, p.Map("a").Type(), "wrong type")
	a.Equal(core.TYPE_STRING, p.Map("a").Stream().Type(), "wrong type")
	a.Equal(core.TYPE_BOOLEAN, p.Map("b").Type(), "wrong type")
}

func TestNewPort__NestedMap(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"},"b":{"type":"map","map":{"a":{"type":"number"}}}}}`)
	p, err := core.NewPort(nil, nil, def, core.DIRECTION_IN, nil)
	a.NoError(err)
	a.NotNil(p)
	a.Equal(core.TYPE_NUMBER, p.Map("a").Type(), "wrong type")
	a.Equal(core.TYPE_MAP, p.Map("b").Type(), "wrong type")
	a.Equal(core.TYPE_NUMBER,p.Map("b").Map("a").Type(), "wrong type")
}

func TestNewPort__Complex(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(
		`{"type":"map","map":{"a":{"type":"stream","stream":{"type":"boolean"}},"b":{"type":"map","map":
{"a":{"type":"stream","stream":{"type":"stream","stream":{"type":"map","map":{"a":{"type":"number"},
"b":{"type":"string"},"c":{"type":"boolean"}}}}},"b":{"type":"string"}}},"c":{"type":"boolean"}}}`)
	p, err := core.NewPort(nil, nil, def, core.DIRECTION_IN, nil)
	a.NoError(err)
	a.Equal(core.TYPE_MAP, p.Type(), "wrong type")
	a.Equal(core.TYPE_STREAM, p.Map("a").Type(), "wrong type")
	a.Equal(core.TYPE_BOOLEAN, p.Map("a").Stream().Type(), "wrong type")
	a.Equal(core.TYPE_MAP, p.Map("b").Type(), "wrong type")
	a.Equal(core.TYPE_STREAM, p.Map("b").Map("a").Type(), "wrong type")
	a.Equal(core.TYPE_STREAM, p.Map("b").Map("a").Stream().Type(), "wrong type")
	a.Equal(core.TYPE_MAP, p.Map("b").Map("a").Stream().Stream().Type(), "wrong type")
	a.Equal(core.TYPE_NUMBER, p.Map("b").Map("a").Stream().Stream().Map("a").Type(), "wrong type")
	a.Equal(core.TYPE_STRING, p.Map("b").Map("a").Stream().Stream().Map("b").Type(), "wrong type")
	a.Equal(core.TYPE_BOOLEAN, p.Map("b").Map("a").Stream().Stream().Map("c").Type(), "wrong type")
	a.Equal(core.TYPE_STRING, p.Map("b").Map("b").Type(), "wrong type")
	a.Equal(core.TYPE_BOOLEAN, p.Map("c").Type(), "wrong type")
}

// Port.Type (6 tests)

func TestPort_Type__Simple__Primitive(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"primitive"}`)
	p, _ := core.NewPort(nil, nil, def, core.DIRECTION_IN, nil)
	a.Equal(core.TYPE_PRIMITIVE, p.Type(), "wrong type")
}

func TestPort_Type__Simple__Trigger(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"trigger"}`)
	p, _ := core.NewPort(nil, nil, def, core.DIRECTION_IN, nil)
	a.Equal(core.TYPE_TRIGGER, p.Type(), "wrong type")
}

func TestPort_Type__Simple__Number(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"number"}`)
	p, _ := core.NewPort(nil, nil, def, core.DIRECTION_IN, nil)
	a.Equal(core.TYPE_NUMBER, p.Type(), "wrong type")
}

func TestPort_Type__Simple__String(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"string"}`)
	p, _ := core.NewPort(nil, nil, def, core.DIRECTION_IN, nil)
	a.Equal(core.TYPE_STRING, p.Type(), "wrong type")
}

func TestPort_Type__Simple__Boolean(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"boolean"}`)
	p, _ := core.NewPort(nil, nil, def, core.DIRECTION_IN, nil)
	a.Equal(core.TYPE_BOOLEAN, p.Type(), "wrong type")
}

func TestPort_Type__Map(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"boolean"}}}`)
	p, _ := core.NewPort(nil, nil, def, core.DIRECTION_IN, nil)
	a.Equal(core.TYPE_MAP, p.Type(), "wrong type")
	a.Equal(core.TYPE_BOOLEAN, p.Map("a").Type(), "wrong type")
}

func TestPort_Type__Stream(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"stream","stream":{"type":"string"}}`)
	p, _ := core.NewPort(nil, nil, def, core.DIRECTION_IN, nil)
	a.Equal(core.TYPE_STREAM, p.Type(), "wrong type")
	a.Equal(core.TYPE_STRING, p.Stream().Type(), "wrong type")
}

// Port.Connect (6 tests)

func TestPort_Connect__Map__KeysNotMatching(t *testing.T) {
	a := assertions.New(t)
	def1 := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := slang.ParsePortDef(`{"type":"map","map":{"b":{"type":"number"}}}`)

	p, _ := core.NewPort(nil, nil, def1, core.DIRECTION_IN, nil)
	q, _ := core.NewPort(nil, nil, def2, core.DIRECTION_OUT, nil)

	err := p.Connect(q)
	a.Error(err)
	a.False(p.Connected(q), "connection not expected")
}

func TestPort_Connect__Map__LeftIsSubsetOfRight(t *testing.T) {
	a := assertions.New(t)
	def1 := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"},"b":{"type":"number"}}}`)

	p, _ := core.NewPort(nil, nil, def1, core.DIRECTION_IN, nil)
	q, _ := core.NewPort(nil, nil, def2, core.DIRECTION_OUT, nil)

	err := p.Connect(q)
	a.Error(err)
	a.False(p.Connected(q), "connection not expected")
}

func TestPort_Connect__Map__RightIsSubsetOfRight(t *testing.T) {
	a := assertions.New(t)
	def1 := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"},"b":{"type":"number"}}}`)
	def2 := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)

	p, _ := core.NewPort(nil, nil, def1, core.DIRECTION_IN, nil)
	q, _ := core.NewPort(nil, nil, def2, core.DIRECTION_OUT, nil)

	err := p.Connect(q)
	a.Error(err)
	a.False(p.Connected(q), "connection not expected")
}

func TestPort_Connect__Map__IncompatibleTypes(t *testing.T) {
	a := assertions.New(t)
	def1 := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := slang.ParsePortDef(`{"type":"number"}`)

	p, _ := core.NewPort(nil, nil, def1, core.DIRECTION_IN, nil)
	q, _ := core.NewPort(nil, nil, def2, core.DIRECTION_OUT, nil)

	err := p.Connect(q)
	a.Error(err)
	a.False(p.Connected(q), "connection not expected")
}

func TestPort_Connect__Map__SameKeys(t *testing.T) {
	a := assertions.New(t)
	def := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)

	p, _ := core.NewPort(nil, nil, def, core.DIRECTION_IN, nil)
	q, _ := core.NewPort(nil, nil, def, core.DIRECTION_OUT, nil)

	err := p.Connect(q)
	a.NoError(err)
	a.True(p.Map("a").Connected(q.Map("a")), "port 'a' must be connected")
	a.False(p.Connected(q), "maps must not be connected")
}

func TestPort_Connect__Map__Subport2Primitive(t *testing.T) {
	a := assertions.New(t)
	def1 := slang.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := slang.ParsePortDef(`{"type":"number"}`)

	p, _ := core.NewPort(nil, nil, def1, core.DIRECTION_IN, nil)
	q, _ := core.NewPort(nil, nil, def2, core.DIRECTION_OUT, nil)

	err := p.Map("a").Connect(q)
	a.NoError(err)
	a.True(p.Map("a").Connected(q), "connection expected")
}
