package tests

import (
	"testing"
	"slang"
	"slang/tests/assertions"
	"slang/core"
)

func TestOperatorDef_SpecifyGenericPorts__NilGenerics(t *testing.T) {
	a := assertions.New(t)
	op := slang.ParseOperatorDef(`{"in": {"type": "number"}, "out": {"type": "number"}}`)
	a.NoError(op.Validate())
	a.NoError(op.SpecifyGenericPorts(nil))
}

func TestOperatorDef_SpecifyGenericPorts__InPortGenerics(t *testing.T) {
	a := assertions.New(t)
	op := slang.ParseOperatorDef(`{"in": {"type": "generic", "generic": "g1"}, "out": {"type": "number"}}`)
	a.NoError(op.Validate())
	a.NoError(op.SpecifyGenericPorts(map[string]*core.PortDef{
		"g1": {
			Type: "boolean",
		},
	}))
	a.Equal("boolean", op.In.Type)
}

func TestOperatorDef_SpecifyGenericPorts__OutPortGenerics(t *testing.T) {
	a := assertions.New(t)
	op := slang.ParseOperatorDef(`{"in": {"type": "number"}, "out": {"type": "generic", "generic": "g1"}}`)
	a.NoError(op.Validate())
	a.NoError(op.SpecifyGenericPorts(map[string]*core.PortDef{
		"g1": {
			Type: "boolean",
		},
	}))
	a.Equal("boolean", op.Out.Type)
}

// TODO: Write more tests for OperatorDef.SpecifyGenericPorts(...)
// TODO: Write tests for PortDef.SpecifyGenericPorts(...)
// TODO: Write tests for OperatorDef.FreeOfGenerics()
// TODO: Write tests for PortDef.FreeOfGenerics()
// TODO: Write tests for PortDef.Copy()