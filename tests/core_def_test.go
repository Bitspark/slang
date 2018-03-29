package tests

import (
	"testing"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/stretchr/testify/require"
	"github.com/Bitspark/slang/pkg/api"
)

// INSTANCE DEFINITION

func TestInstanceDef_Validate__FailsMissingName(t *testing.T) {
	a := assertions.New(t)
	_, err := validateJSONInstanceDef(`{
		"operator": "opr"
	}`)
	a.Error(err)
}

func TestInstanceDef_Validate__FailsSpacesInName(t *testing.T) {
	a := assertions.New(t)
	_, err := validateJSONInstanceDef(`{
		"operator": "opr",
		"name":"fun 4 ever",
	}`)
	a.Error(err)
}

func TestInstanceDef_Validate__FailsMissingOperator(t *testing.T) {
	a := assertions.New(t)
	_, err := validateJSONInstanceDef(`{
		"name":"oprInstance"
	}`)
	a.Error(err)
}

func TestInstanceDef_Validate__Succeeds(t *testing.T) {
	a := assertions.New(t)
	ins, err := validateJSONInstanceDef(`{
		"operator": "opr",
		"name":"oprInstance"
	}`)
	a.NoError(err)
	a.True(ins.Valid())
}

// OPERATOR DEFINITION

func TestOperatorDef_Validate__FailsPortMustBeDefined_In(t *testing.T) {
	a := assertions.New(t)
	_, err := validateJSONOperatorDef(`{
		"services": {"main": {"out": {"type":"number"}}}
	}`)
	a.Error(err)
}

func TestOperatorDef_Validate__FailsPortMustBeDefined_Out(t *testing.T) {
	a := assertions.New(t)
	_, err := validateJSONOperatorDef(`{
		"services": {"main": {"in": {"type":"number"}}}
	}`)
	a.Error(err)
}

func TestOperatorDef_Validate__Succeeds(t *testing.T) {
	a := assertions.New(t)
	oDef, err := validateJSONOperatorDef(`{
		"services": {"main": {
		"in": {
			"type": "number"
		},
		"out": {
			"type": "number"
		}}},
		"operators": [
			{
				"operator": "builtin_Adder",
				"name": "add"
			}
		],
		"connections": {
			"(": ["(add"],
			"add)": [")"]
		}
	}`)
	a.NoError(err)
	a.True(oDef.Valid())
}

func TestOperatorDef_SpecifyGenericPorts__NilGenerics(t *testing.T) {
	a := assertions.New(t)
	op, _ := api.ParseJSONOperatorDef(`{"services": {"` + core.MAIN_SERVICE + `": {"in": {"type": "number"}, "out": {"type": "number"}}}`)
	require.NoError(t, op.Validate())
	a.NoError(op.SpecifyGenericPorts(nil))
}

func TestOperatorDef_SpecifyGenericPorts__InPortGenerics(t *testing.T) {
	a := assertions.New(t)
	op, _ := api.ParseJSONOperatorDef(`{"services": {"` + core.MAIN_SERVICE + `": {"in": {"type": "generic", "generic": "g1"}, "out": {"type": "number"}}}}`)
	require.NoError(t, op.Validate())
	a.NoError(op.SpecifyGenericPorts(map[string]*core.TypeDef{
		"g1": {
			Type: "boolean",
		},
	}))
	a.Equal("boolean", op.ServiceDefs[core.MAIN_SERVICE].In.Type)
}

func TestOperatorDef_SpecifyGenericPorts__OutPortGenerics(t *testing.T) {
	a := assertions.New(t)
	op, _ := api.ParseJSONOperatorDef(`{"services": {"` + core.MAIN_SERVICE + `": {"in": {"type": "number"}, "out": {"type": "generic", "generic": "g1"}}}}`)
	require.NoError(t, op.Validate())
	a.NoError(op.SpecifyGenericPorts(map[string]*core.TypeDef{
		"g1": {
			Type: "boolean",
		},
	}))
	a.Equal("boolean", op.ServiceDefs[core.MAIN_SERVICE].Out.Type)
}

func TestOperatorDef_SpecifyGenericPorts__GenericPortsGenerics(t *testing.T) {
	a := assertions.New(t)
	op, _ := api.ParseJSONOperatorDef(
		`{"services": {"` + core.MAIN_SERVICE + `": {"in": {"type": "number"}, "out": {"type": "number"}}}, "operators": {"test": {"operator": "fork", "generics": {"itemType": {"type": "generic", "generic": "g1"}}}}}`)
	require.NoError(t, op.Validate())
	a.NoError(op.SpecifyGenericPorts(map[string]*core.TypeDef{
		"g1": {
			Type: "boolean",
		},
	}))
	a.Equal("boolean", op.InstanceDefs[0].Generics["itemType"].Type)
}

func TestOperatorDef_SpecifyGenericPorts__DifferentIdentifier(t *testing.T) {
	a := assertions.New(t)
	op, _ := api.ParseJSONOperatorDef(`{"services": {"` + core.MAIN_SERVICE + `": {"in": {"type": "generic", "generic": "g1"}, "out": {"type": "number"}}}}`)
	require.NoError(t, op.Validate())
	a.NoError(op.SpecifyGenericPorts(map[string]*core.TypeDef{
		"g2": {
			Type: "boolean",
		},
	}))
	a.Equal("generic", op.ServiceDefs[core.MAIN_SERVICE].In.Type)
}

func TestOperatorDef_GenericsSpecified__InPortGenerics(t *testing.T) {
	a := assertions.New(t)
	op, _ := api.ParseJSONOperatorDef(`{"services": {"` + core.MAIN_SERVICE + `": {"in": {"type": "generic", "generic": "t1"}, "out": {"type": "number"}}}}`)
	require.NoError(t, op.Validate())
	a.Error(op.GenericsSpecified())
}

func TestOperatorDef_GenericsSpecified__InPortNoGenerics(t *testing.T) {
	a := assertions.New(t)
	op, _ := api.ParseJSONOperatorDef(`{"services": {"` + core.MAIN_SERVICE + `": {"in": {"type": "string"}, "out": {"type": "number"}}}}`)
	require.NoError(t, op.Validate())
	a.NoError(op.GenericsSpecified())
}

func TestOperatorDef_GenericsSpecified__OutPortGenerics(t *testing.T) {
	a := assertions.New(t)
	op, _ := api.ParseJSONOperatorDef(`{"services": {"` + core.MAIN_SERVICE + `": {"in": {"type": "number"}, "out": {"type": "generic", "generic": "t1"}}}}`)
	require.NoError(t, op.Validate())
	a.Error(op.GenericsSpecified())
}

func TestOperatorDef_GenericsSpecified__OutPortNoGenerics(t *testing.T) {
	a := assertions.New(t)
	op, _ := api.ParseJSONOperatorDef(`{"services": {"` + core.MAIN_SERVICE + `": {"in": {"type": "number"}, "out": {"type": "string"}}}}`)
	require.NoError(t, op.Validate())
	a.NoError(op.GenericsSpecified())
}

func TestOperatorDef_GenericsSpecified__GenericPortsGenerics(t *testing.T) {
	a := assertions.New(t)
	op, _ := api.ParseJSONOperatorDef(
		`{"services": {"` + core.MAIN_SERVICE + `": {"in": {"type": "number"}, "out": {"type": "number"}}}, "operators": {"test": {"operator": "fork", "generics": {"itemType": {"type": "generic", "generic": "g1"}}}}}`)
	require.NoError(t, op.Validate())
	a.Error(op.GenericsSpecified())
}

func TestOperatorDef_GenericsSpecified__GenericPortsNoGenerics(t *testing.T) {
	a := assertions.New(t)
	op, _ := api.ParseJSONOperatorDef(
		`{"services": {"` + core.MAIN_SERVICE + `": {"in": {"type": "number"}, "out": {"type": "number"}}}, "operators": [{"name": "test", "operator": "fork", "generics": {"itemType": {"type": "number"}}}]}`)
	require.NoError(t, op.Validate())
	a.NoError(op.GenericsSpecified())
}

// PORT DEFINITION

func TestPortDef_Copy__Simple(t *testing.T) {
	a := assertions.New(t)
	pd := core.TypeDef{Type: "number"}
	require.NoError(t, pd.Validate())
	pdCpy := pd.Copy()
	require.NoError(t, pdCpy.Validate())
	a.Equal("number", pdCpy.Type)
}

func TestPortDef_Copy__Stream(t *testing.T) {
	a := assertions.New(t)
	pd := core.TypeDef{Type: "stream", Stream: &core.TypeDef{Type: "string"}}
	require.NoError(t, pd.Validate())
	pdCpy := pd.Copy()
	require.NoError(t, pdCpy.Validate())
	a.Equal("string", pdCpy.Stream.Type)
	a.False(pd.Stream == pdCpy.Stream)
}

func TestPortDef_Copy__Map(t *testing.T) {
	a := assertions.New(t)
	pd := core.TypeDef{Type: "map", Map: map[string]*core.TypeDef{"a": {Type: "string"}}}
	require.NoError(t, pd.Validate())
	pdCpy := pd.Copy()
	require.NoError(t, pdCpy.Validate())
	a.Equal("string", pdCpy.Map["a"].Type)
	a.False(pd.Map["a"] == pdCpy.Map["a"])
}

func TestPortDef_SpecifyGenericPorts__Simple(t *testing.T) {
	a := assertions.New(t)
	pd := core.TypeDef{Type: "generic", Generic: "t1"}
	require.NoError(t, pd.Validate())
	a.NoError(pd.SpecifyGenerics(map[string]*core.TypeDef{
		"t1": {Type: "number"},
	}))
	a.Equal("number", pd.Type)
}

func TestPortDef_SpecifyGenericPorts__DifferentIdentifier(t *testing.T) {
	a := assertions.New(t)
	pd := core.TypeDef{Type: "generic", Generic: "t1"}
	require.NoError(t, pd.Validate())
	a.NoError(pd.SpecifyGenerics(map[string]*core.TypeDef{
		"t2": {Type: "number"},
	}))
	a.Equal("generic", pd.Type)
}

func TestPortDef_SpecifyGenericPorts__MultipleIdentifiers(t *testing.T) {
	a := assertions.New(t)
	pd := core.TypeDef{
		Type: "map",
		Map:  map[string]*core.TypeDef{"a": {Type: "generic", Generic: "t1"}, "b": {Type: "generic", Generic: "t2"}},
	}
	require.NoError(t, pd.Validate())
	a.NoError(pd.SpecifyGenerics(map[string]*core.TypeDef{
		"t1": {Type: "number"},
	}))
	a.Equal("number", pd.Map["a"].Type)
	a.Equal("generic", pd.Map["b"].Type)
}

func TestPortDef_SpecifyGenericPorts__Stream(t *testing.T) {
	a := assertions.New(t)
	pd := core.TypeDef{Type: "stream", Stream: &core.TypeDef{Type: "generic", Generic: "t1"}}
	require.NoError(t, pd.Validate())
	a.NoError(pd.SpecifyGenerics(map[string]*core.TypeDef{
		"t1": {Type: "number"},
	}))
	a.Equal("number", pd.Stream.Type)
}

func TestPortDef_SpecifyGenericPorts__Map(t *testing.T) {
	a := assertions.New(t)
	pd := core.TypeDef{Type: "map", Map: map[string]*core.TypeDef{"a": {Type: "generic", Generic: "t1"}}}
	require.NoError(t, pd.Validate())
	a.NoError(pd.SpecifyGenerics(map[string]*core.TypeDef{
		"t1": {Type: "number"},
	}))
	a.Equal("number", pd.Map["a"].Type)
}

func TestPortDef_GenericsSpecified__SimpleGeneric(t *testing.T) {
	a := assertions.New(t)
	pd := core.TypeDef{Type: "generic", Generic: "t1"}
	require.NoError(t, pd.Validate())
	a.Error(pd.GenericsSpecified())
}

func TestPortDef_GenericsSpecified__SimpleNoGeneric(t *testing.T) {
	a := assertions.New(t)
	pd := core.TypeDef{Type: "number"}
	require.NoError(t, pd.Validate())
	a.NoError(pd.GenericsSpecified())
}

func TestPortDef_GenericsSpecified__StreamGenerics(t *testing.T) {
	a := assertions.New(t)
	pd := core.TypeDef{Type: "stream", Stream: &core.TypeDef{Type: "generic", Generic: "t1"}}
	require.NoError(t, pd.Validate())
	a.Error(pd.GenericsSpecified())
}

func TestPortDef_GenericsSpecified__StreamNoGenerics(t *testing.T) {
	a := assertions.New(t)
	pd := core.TypeDef{Type: "stream", Stream: &core.TypeDef{Type: "number"}}
	require.NoError(t, pd.Validate())
	a.NoError(pd.GenericsSpecified())
}

func TestPortDef_GenericsSpecified__MapGenerics(t *testing.T) {
	a := assertions.New(t)
	pd := core.TypeDef{Type: "map", Map: map[string]*core.TypeDef{"a": {Type: "number"}}}
	require.NoError(t, pd.Validate())
	a.NoError(pd.GenericsSpecified())
}

// PROPERTY PARSING

func makeProps() (map[string]*core.TypeDef, core.Properties) {
	propDefs := make(map[string]*core.TypeDef)
	props := make(core.Properties)
	propDefs["strvar"] = &core.TypeDef{Type: "string"}
	propDefs["numvar"] = &core.TypeDef{Type: "number"}
	propDefs["boolvar"] = &core.TypeDef{Type: "boolean"}
	propDefs["arrvar1"] = &core.TypeDef{Type: "stream", Stream: &core.TypeDef{Type: "string"}}
	propDefs["arrvar2"] = &core.TypeDef{Type: "stream", Stream: &core.TypeDef{Type: "number"}}
	return propDefs, props
}

func TestExpandExpression__Empty(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)
	propDefs, props := makeProps()
	parts, err := core.ExpandExpression("", props, propDefs)
	r.NoError(err)
	a.Equal([]string{""}, parts)
}

func TestExpandExpression__SingleString(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)
	propDefs, props := makeProps()
	parts, err := core.ExpandExpression("test", props, propDefs)
	r.NoError(err)
	a.Equal([]string{"test"}, parts)
}

func TestExpandExpression__String1(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)
	propDefs, props := makeProps()
	props["strvar"] = "testval"
	parts, err := core.ExpandExpression("{$strvar}", props, propDefs)
	r.NoError(err)
	a.Equal([]string{"testval"}, parts)
}

func TestExpandExpression__StringAndNumber(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)
	propDefs, props := makeProps()
	props["strvar"] = "testval"
	props["numvar"] = 12
	parts, err := core.ExpandExpression("{$strvar}_{$numvar}", props, propDefs)
	r.NoError(err)
	a.Equal([]string{"testval_12"}, parts)
}

func TestExpandExpression__Array1(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)
	propDefs, props := makeProps()
	props["arrvar1"] = []interface{}{"a", "b", "c"}
	parts, err := core.ExpandExpression("val_{$arrvar1}_end", props, propDefs)
	r.NoError(err)
	a.Equal([]string{"val_a_end", "val_b_end", "val_c_end"}, parts)
}

func TestExpandExpression__ArrayAndBoolean(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)
	propDefs, props := makeProps()
	props["arrvar1"] = []interface{}{"a", "b", "c"}
	props["boolvar"] = true
	parts, err := core.ExpandExpression("{$arrvar1}_{$boolvar}", props, propDefs)
	r.NoError(err)
	a.Equal([]string{"a_true", "b_true", "c_true"}, parts)
}

func TestExpandExpression__ArrayCross1(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)
	propDefs, props := makeProps()
	props["arrvar1"] = []interface{}{"a", "b", "c"}
	props["arrvar2"] = []interface{}{1, 2}
	parts, err := core.ExpandExpression("{$arrvar1}_{$arrvar2}", props, propDefs)
	r.NoError(err)
	a.Equal([]string{"a_1", "b_1", "c_1", "a_2", "b_2", "c_2"}, parts)
}

func TestExpandExpression__ArrayCross2(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)
	propDefs, props := makeProps()
	props["arrvar1"] = []interface{}{"a", "b", "c"}
	parts, err := core.ExpandExpression("{$arrvar1}_{$arrvar1}", props, propDefs)
	r.NoError(err)
	a.Equal([]string{"a_a", "b_a", "c_a", "a_b", "b_b", "c_b", "a_c", "b_c", "c_c"}, parts)
}
