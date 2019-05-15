package tests

import (
	"testing"
	"time"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
)

// TypeDef.Validate (11 tests)

func TestTypeDef_Validate__InvalidTypeInDefinition(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"gfdhgfd"}`)
	p, err := core.NewPort(nil, nil, def, core.DIRECTION_IN)
	a.Error(err)
	a.Nil(p)
}

func TestTypeDef_Validate__Stream__StreamNotPresent(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"stream"}`)
	a.Error(def.Validate())
	a.False(def.Valid(), "should not be valid")
}

func TestTypeDef_Validate__Stream__NilStream(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"stream","stream":null}`)
	a.Error(def.Validate())
	a.False(def.Valid(), "should not be valid")
}

func TestTypeDef_Validate__Stream__EmptyStream(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"stream","stream":{}}`)
	a.Error(def.Validate())
	a.False(def.Valid(), "should not be valid")
}

func TestTypeDef_Validate__Stream__InvalidTypeInDefinition(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"stream","stream":{"type":"hgfdh"}}`)
	a.Error(def.Validate())
	a.False(def.Valid(), "should not be valid")
}

func TestTypeDef_Validate__Map__NoMapEntries(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"map"}`)
	a.NoError(def.Validate())
	a.True(def.Valid(), "should be valid")
}

func TestTypeDef_Validate__Map__NilMap(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"map","map":null}`)
	a.NoError(def.Validate())
	a.True(def.Valid(), "should be valid")
}

func TestTypeDef_Validate__Map__EmptyMap(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"map","map":{}}`)
	a.NoError(def.Validate())
	a.True(def.Valid(), "should be valid")
}

func TestTypeDef_Validate__Map__NullEntry(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"map","map":{"a":null}}`)
	a.Error(def.Validate())
	a.False(def.Valid(), "should not be valid")
}

func TestTypeDef_Validate__Map__EmptyEntry(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"map","map":{"a":{}}}`)
	a.Error(def.Validate())
	a.False(def.Valid(), "should not be valid")
}

func TestTypeDef_Validate__Map__InvalidTypeInDefinition(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"map","map":{"a":{"type":"gfgfd"}}}`)
	a.Error(def.Validate())
	a.False(def.Valid(), "should not be valid")
}

func TestTypeDef_Validate__Generic__IdentifierMissing(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"generic"}`)
	a.Error(def.Validate())
	a.False(def.Valid(), "should not be valid")
}

func TestTypeDef_Validate__Generic__Correct(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"generic", "generic":"id1"}`)
	a.NoError(def.Validate())
	a.True(def.Valid(), "should be valid")
}

// core.NewPort (10 tests)

func TestNewPort__InvalidDefinition(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"bcvbvcbvc"}`)
	p, err := core.NewPort(nil, nil, def, 0)
	a.Error(err)
	a.Nil(p)
}

func TestNewPort__Number__NoDirectionGiven(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"number"}`)
	p, err := core.NewPort(nil, nil, def, 0)
	a.Error(err)
	a.Nil(p)
}

func TestNewPort__Number__WrongDirectionGiven(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"number"}`)
	p, err := core.NewPort(nil, nil, def, 3)
	a.Error(err)
	a.Nil(p)
}

func TestNewPort__Stream(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"stream","stream":{"type":"number"}}`)
	p, err := core.NewPort(nil, nil, def, core.DIRECTION_IN)
	a.NoError(err)
	a.Equal(core.TYPE_STREAM, p.Type(), "wrong type")
	a.Equal(core.TYPE_NUMBER, p.Stream().Type(), "wrong type")
}

func TestNewPort__Number(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"number"}`)
	p, err := core.NewPort(nil, nil, def, core.DIRECTION_IN)

	a.NoError(err)
	a.NotNil(p)
	a.Equal(core.TYPE_NUMBER, p.Type(), "wrong type")
	a.Equal(core.DIRECTION_IN, p.Direction(), "wrong direction")
}

func TestNewPort__Map(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	p, err := core.NewPort(nil, nil, def, core.DIRECTION_IN)
	a.NoError(err)
	a.Equal(core.TYPE_MAP, p.Type(), "wrong type")
	a.Equal(core.TYPE_NUMBER, p.Map("a").Type(), "wrong type")
}

func TestNewPort__NestedStreams(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"stream","stream":{"type":"stream","stream":{"type":"string"}}}`)
	p, err := core.NewPort(nil, nil, def, core.DIRECTION_IN)
	a.NoError(err)
	a.Equal(core.TYPE_STREAM, p.Type(), "wrong type")
	a.Equal(core.TYPE_STREAM, p.Stream().Type(), "wrong type")
	a.Equal(core.TYPE_STRING, p.Stream().Stream().Type(), "wrong type")
}

func TestNewPort__MapStream(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(
		`{"type":"map","map":{"a":{"type":"stream","stream":{"type":"string"}},"b":{"type":"boolean"}}}`)
	p, err := core.NewPort(nil, nil, def, core.DIRECTION_IN)
	a.NoError(err)
	a.Equal(core.TYPE_MAP, p.Type(), "wrong type")
	a.Equal(core.TYPE_STREAM, p.Map("a").Type(), "wrong type")
	a.Equal(core.TYPE_STRING, p.Map("a").Stream().Type(), "wrong type")
	a.Equal(core.TYPE_BOOLEAN, p.Map("b").Type(), "wrong type")
}

func TestNewPort__NestedMap(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"map","map":{"a":{"type":"number"},"b":{"type":"map","map":{"a":{"type":"number"}}}}}`)
	p, err := core.NewPort(nil, nil, def, core.DIRECTION_IN)
	a.NoError(err)
	a.NotNil(p)
	a.Equal(core.TYPE_NUMBER, p.Map("a").Type(), "wrong type")
	a.Equal(core.TYPE_MAP, p.Map("b").Type(), "wrong type")
	a.Equal(core.TYPE_NUMBER, p.Map("b").Map("a").Type(), "wrong type")
}

func TestNewPort__Complex(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(
		`{"type":"map","map":{"a":{"type":"stream","stream":{"type":"boolean"}},"b":{"type":"map","map":
{"a":{"type":"stream","stream":{"type":"stream","stream":{"type":"map","map":{"a":{"type":"number"},
"b":{"type":"string"},"c":{"type":"boolean"}}}}},"b":{"type":"string"}}},"c":{"type":"boolean"}}}`)
	p, err := core.NewPort(nil, nil, def, core.DIRECTION_IN)
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
	def := core.ParseTypeDef(`{"type":"primitive"}`)
	p, _ := core.NewPort(nil, nil, def, core.DIRECTION_IN)
	a.Equal(core.TYPE_PRIMITIVE, p.Type(), "wrong type")
}

func TestPort_Type__Simple__Trigger(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"trigger"}`)
	p, _ := core.NewPort(nil, nil, def, core.DIRECTION_IN)
	a.Equal(core.TYPE_TRIGGER, p.Type(), "wrong type")
}

func TestPort_Type__Simple__Number(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"number"}`)
	p, _ := core.NewPort(nil, nil, def, core.DIRECTION_IN)
	a.Equal(core.TYPE_NUMBER, p.Type(), "wrong type")
}

func TestPort_Type__Simple__String(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"string"}`)
	p, _ := core.NewPort(nil, nil, def, core.DIRECTION_IN)
	a.Equal(core.TYPE_STRING, p.Type(), "wrong type")
}

func TestPort_Type__Simple__Boolean(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"boolean"}`)
	p, _ := core.NewPort(nil, nil, def, core.DIRECTION_IN)
	a.Equal(core.TYPE_BOOLEAN, p.Type(), "wrong type")
}

func TestPort_Type__Map(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"map","map":{"a":{"type":"boolean"}}}`)
	p, _ := core.NewPort(nil, nil, def, core.DIRECTION_IN)
	a.Equal(core.TYPE_MAP, p.Type(), "wrong type")
	a.Equal(core.TYPE_BOOLEAN, p.Map("a").Type(), "wrong type")
}

func TestPort_Type__Stream(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"stream","stream":{"type":"string"}}`)
	p, _ := core.NewPort(nil, nil, def, core.DIRECTION_IN)
	a.Equal(core.TYPE_STREAM, p.Type(), "wrong type")
	a.Equal(core.TYPE_STRING, p.Stream().Type(), "wrong type")
}

// Port.WalkPrimitives

func TestPort_WalkPrimitives__Simple_Primitive(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"primitive"}`)
	p, _ := core.NewPort(nil, nil, def, core.DIRECTION_IN)

	ports := make(chan string)
	portNames := make([]string, 0)

	p.WalkPrimitivePorts(func(p *core.Port) {
		ports <- p.Name()
	})

outer:
	for {
		select {
		case portName := <-ports:
			portNames = append(portNames, portName)
		case <-time.After(1 * time.Millisecond):
			break outer
		}
	}

	a.Len(portNames, 1)
	a.Equal(portNames[0], "(")
}

func TestPort_WalkPrimitives__Stream(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"stream","stream":{"type":"string"}}`)
	p, _ := core.NewPort(nil, nil, def, core.DIRECTION_IN)

	ports := make(chan string)
	portNames := make([]string, 0)

	p.WalkPrimitivePorts(func(p *core.Port) {
		ports <- p.Name()
	})

outer:
	for {
		select {
		case portName := <-ports:
			portNames = append(portNames, portName)
		case <-time.After(1 * time.Millisecond):
			break outer
		}
	}

	a.Len(portNames, 1)
	a.Equal(portNames[0], "~(")
}

func TestPort_WalkPrimitives__Stream_Map(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"map","map":{"p":{"type":"primitive"},"sp":{"type":"stream","stream":{"type":"primitive"}}}}`)
	p, _ := core.NewPort(nil, nil, def, core.DIRECTION_IN)

	ports := make(chan string)
	portNames := make([]string, 0)

	p.WalkPrimitivePorts(func(p *core.Port) {
		ports <- p.Name()
	})

outer:
	for {
		select {
		case portName := <-ports:
			portNames = append(portNames, portName)
		case <-time.After(1 * time.Millisecond):
			break outer
		}
	}

	a.Len(portNames, 2)
	a.Contains(portNames, "p(")
	a.Contains(portNames, "sp.~(")
}

// Port.Connect (6 tests)

func TestPort_Connect__Map__KeysNotMatching(t *testing.T) {
	a := assertions.New(t)
	def1 := core.ParseTypeDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := core.ParseTypeDef(`{"type":"map","map":{"b":{"type":"number"}}}`)

	p, _ := core.NewPort(nil, nil, def1, core.DIRECTION_IN)
	q, _ := core.NewPort(nil, nil, def2, core.DIRECTION_OUT)

	err := p.Connect(q)
	a.Error(err)
	a.False(p.Connected(q), "connection not expected")
}

func TestPort_Connect__Map__LeftIsSubsetOfRight(t *testing.T) {
	a := assertions.New(t)
	def1 := core.ParseTypeDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := core.ParseTypeDef(`{"type":"map","map":{"a":{"type":"number"},"b":{"type":"number"}}}`)

	p, _ := core.NewPort(nil, nil, def1, core.DIRECTION_IN)
	q, _ := core.NewPort(nil, nil, def2, core.DIRECTION_OUT)

	err := p.Connect(q)
	a.Error(err)
	a.False(p.Connected(q), "connection not expected")
}

func TestPort_Connect__Map__RightIsSubsetOfRight(t *testing.T) {
	a := assertions.New(t)
	def1 := core.ParseTypeDef(`{"type":"map","map":{"a":{"type":"number"},"b":{"type":"number"}}}`)
	def2 := core.ParseTypeDef(`{"type":"map","map":{"a":{"type":"number"}}}`)

	p, _ := core.NewPort(nil, nil, def1, core.DIRECTION_IN)
	q, _ := core.NewPort(nil, nil, def2, core.DIRECTION_OUT)

	err := p.Connect(q)
	a.Error(err)
	a.False(p.Connected(q), "connection not expected")
}

func TestPort_Connect__Map__IncompatibleTypes(t *testing.T) {
	a := assertions.New(t)
	def1 := core.ParseTypeDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := core.ParseTypeDef(`{"type":"number"}`)

	p, _ := core.NewPort(nil, nil, def1, core.DIRECTION_IN)
	q, _ := core.NewPort(nil, nil, def2, core.DIRECTION_OUT)

	err := p.Connect(q)
	a.Error(err)
	a.False(p.Connected(q), "connection not expected")
}

func TestPort_Connect__Map__SameKeys(t *testing.T) {
	a := assertions.New(t)
	def := core.ParseTypeDef(`{"type":"map","map":{"a":{"type":"number"}}}`)

	p, _ := core.NewPort(nil, nil, def, core.DIRECTION_IN)
	q, _ := core.NewPort(nil, nil, def, core.DIRECTION_OUT)

	err := p.Connect(q)
	a.NoError(err)
	a.True(p.Map("a").Connected(q.Map("a")), "port 'a' must be connected")
	a.False(p.Connected(q), "maps must not be connected")
}

func TestPort_Connect__Map__Subport2Primitive(t *testing.T) {
	a := assertions.New(t)
	def1 := core.ParseTypeDef(`{"type":"map","map":{"a":{"type":"number"}}}`)
	def2 := core.ParseTypeDef(`{"type":"number"}`)

	p, _ := core.NewPort(nil, nil, def1, core.DIRECTION_IN)
	q, _ := core.NewPort(nil, nil, def2, core.DIRECTION_OUT)

	err := p.Map("a").Connect(q)
	a.NoError(err)
	a.True(p.Map("a").Connected(q), "connection expected")
}
