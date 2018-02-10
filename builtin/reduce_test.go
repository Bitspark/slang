package builtin

import (
	"testing"
	"slang/tests/assertions"
	"slang/core"
	"github.com/stretchr/testify/require"
)

func TestBuiltin_Reduce__CreatorFuncIsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocReduce := getBuiltinCfg("reduce")
	a.NotNil(ocReduce)
}

func TestBuiltin_Reduce__NoGenerics(t *testing.T) {
	a := assertions.New(t)

	_, err := MakeOperator(core.InstanceDef{Operator: "reduce"})
	a.Error(err)
}

func TestBuiltin_Reduce__InPorts(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := MakeOperator(core.InstanceDef{Operator: "reduce", Generics: map[string]*core.PortDef{"itemType": {Type: "number"}}})
	r.NoError(err)

	a.Equal(core.TYPE_MAP, o.In().Type())
	a.Equal(core.TYPE_STREAM, o.In().Map("items").Type())
	a.Equal(core.TYPE_STREAM, o.In().Map("pool").Type())

	// Item type
	itemType := core.TYPE_NUMBER
	a.Equal(itemType, o.In().Map("items").Stream().Type())
	a.Equal(itemType, o.In().Map("pool").Stream().Type())

	o, err = MakeOperator(core.InstanceDef{Operator: "reduce", Generics: map[string]*core.PortDef{"itemType": {Type: "string"}}})
	r.NoError(err)

	// Item type
	itemType = core.TYPE_STRING
	a.Equal(itemType, o.In().Map("items").Stream().Type())
	a.Equal(itemType, o.In().Map("pool").Stream().Type())
	r.NoError(err)
}

func TestBuiltin_Reduce__OutPorts(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := MakeOperator(core.InstanceDef{Operator: "reduce", Generics: map[string]*core.PortDef{"itemType": {Type: "number"}}})
	r.NoError(err)

	a.Equal(core.TYPE_MAP, o.Out().Type())
	a.Equal(core.TYPE_NUMBER, o.Out().Map("result").Type())
	a.Equal(core.TYPE_STREAM, o.Out().Map("selection").Type())
	a.Equal(core.TYPE_MAP, o.Out().Map("selection").Stream().Type())

	// Item type
	itemType := core.TYPE_NUMBER
	a.Equal(itemType, o.Out().Map("selection").Stream().Map("a").Type())
	a.Equal(itemType, o.Out().Map("selection").Stream().Map("b").Type())

	o, err = MakeOperator(core.InstanceDef{Operator: "reduce", Generics: map[string]*core.PortDef{"itemType": {Type: "string"}}})
	r.NoError(err)

	// Item type
	itemType = core.TYPE_STRING
	a.Equal(itemType, o.Out().Map("selection").Stream().Map("a").Type())
	a.Equal(itemType, o.Out().Map("selection").Stream().Map("b").Type())
}

func TestBuiltin_Reduce__PassMarkers(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := MakeOperator(core.InstanceDef{Operator: "reduce", Generics: map[string]*core.PortDef{"itemType": {Type: "number"}}})
	r.NoError(err)

	o.Out().Map("result").Bufferize()
	o.Out().Map("selection").Stream().Bufferize()
	o.Start()

	bos := core.BOS{}
	eos := core.BOS{}
	o.In().Map("items").Stream().Push(bos)
	o.In().Map("items").Stream().Push(eos)
	o.In().Map("pool").Stream().Push(bos)
	o.In().Map("pool").Stream().Push(eos)

	a.PortPushes([]interface{}{bos, eos}, o.Out().Map("result"))
}

func TestBuiltin_Reduce__SelectionFromItemsEmpty(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := MakeOperator(core.InstanceDef{
		Operator: "reduce",
		Generics: map[string]*core.PortDef{"itemType": {Type: "string"}},
		Properties: map[string]interface{}{"emptyValue": "empty"},
	})
	r.NoError(err)

	o.Out().Map("result").Bufferize()
	o.Out().Map("selection").Stream().Map("a").Bufferize()
	o.Out().Map("selection").Stream().Map("b").Bufferize()
	o.Start()

	o.In().Map("items").Push([]interface{}{})

	i := o.Out().Map("result").Pull()
	a.Equal("empty", i)
}

func TestBuiltin_Reduce__SelectionFromItemsSingle(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := MakeOperator(core.InstanceDef{Operator: "reduce", Generics: map[string]*core.PortDef{"itemType": {Type: "number"}}})
	r.NoError(err)

	o.Out().Map("result").Bufferize()
	o.Out().Map("selection").Stream().Map("a").Bufferize()
	o.Out().Map("selection").Stream().Map("b").Bufferize()
	o.Start()

	o.In().Map("items").Push([]interface{}{123.0})

	i := o.Out().Map("result").Pull()
	a.Equal(123.0, i)
}

func TestBuiltin_Reduce__SelectionFromItemsMultiple(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := MakeOperator(core.InstanceDef{Operator: "reduce", Generics: map[string]*core.PortDef{"itemType": {Type: "number"}}})
	r.NoError(err)

	o.Out().Map("result").Bufferize()
	o.Out().Map("selection").Stream().Map("a").Bufferize()
	o.Out().Map("selection").Stream().Map("b").Bufferize()
	o.Start()

	o.In().Map("items").Push([]interface{}{1.0, 2.0})
	o.In().Map("pool").Push([]interface{}{3.0})

	i := o.Out().Map("selection").Pull()
	a.Equal([]interface{}{map[string]interface{}{"a": 1.0, "b": 2.0}}, i)

	i = o.Out().Map("result").Pull()
	a.Equal(3.0, i)
}

func TestBuiltin_Reduce__SelectionFromPool(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := MakeOperator(core.InstanceDef{Operator: "reduce", Generics: map[string]*core.PortDef{"itemType": {Type: "number"}}})
	r.NoError(err)

	o.Out().Map("result").Bufferize()
	o.Out().Map("selection").Stream().Map("a").Bufferize()
	o.Out().Map("selection").Stream().Map("b").Bufferize()
	o.Start()

	o.In().Map("items").Push([]interface{}{1.0, 2.0})
	o.In().Map("pool").Push([]interface{}{3.0, 4.0, 5.0, 6.0})

	i := o.Out().Map("selection").Pull()
	a.Equal([]interface{}{map[string]interface{}{"a": 1.0, "b": 2.0}}, i)

	i = o.Out().Map("selection").Pull()
	a.Equal([]interface{}{
		map[string]interface{}{"a": 3.0, "b": 4.0},
		map[string]interface{}{"a": 5.0, "b": 6.0},
	}, i)
}

func TestBuiltin_Reduce__MixedSelection1(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := MakeOperator(core.InstanceDef{Operator: "reduce", Generics: map[string]*core.PortDef{"itemType": {Type: "number"}}})
	r.NoError(err)

	o.Out().Map("result").Bufferize()
	o.Out().Map("selection").Stream().Map("a").Bufferize()
	o.Out().Map("selection").Stream().Map("b").Bufferize()
	o.Start()

	o.In().Map("items").Push([]interface{}{1.0, 2.0, 3.0})
	o.In().Map("pool").Push([]interface{}{4.0})

	i := o.Out().Map("selection").Pull()
	a.Equal([]interface{}{map[string]interface{}{"a": 1.0, "b": 2.0}}, i)
	i = o.Out().Map("selection").Pull()
	a.Equal([]interface{}{map[string]interface{}{"a": 3.0, "b": 4.0}}, i)
}

func TestBuiltin_Reduce__MixedSelection2(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := MakeOperator(core.InstanceDef{Operator: "reduce", Generics: map[string]*core.PortDef{"itemType": {Type: "number"}}})
	r.NoError(err)

	o.Out().Map("result").Bufferize()
	o.Out().Map("selection").Stream().Map("a").Bufferize()
	o.Out().Map("selection").Stream().Map("b").Bufferize()
	o.Start()

	o.In().Map("items").Push([]interface{}{1.0, 2.0, 3.0})
	o.In().Map("pool").Push([]interface{}{4.0, 5.0, 6.0})
	o.In().Map("pool").Push([]interface{}{7.0, 8.0, 9.0})
	o.In().Map("pool").Push([]interface{}{10.0})
	o.In().Map("pool").Push([]interface{}{11.0})

	i := o.Out().Map("selection").Pull()
	a.Equal([]interface{}{map[string]interface{}{"a": 1.0, "b": 2.0}}, i)

	i = o.Out().Map("selection").Pull()
	a.Equal([]interface{}{
		map[string]interface{}{"a": 3.0, "b": 4.0},
		map[string]interface{}{"a": 5.0, "b": 6.0},
	}, i)

	i = o.Out().Map("selection").Pull()
	a.Equal([]interface{}{
		map[string]interface{}{"a": 7.0, "b": 8.0},
	}, i)

	i = o.Out().Map("selection").Pull()
	a.Equal([]interface{}{
		map[string]interface{}{"a": 9.0, "b": 10.0},
	}, i)

	i = o.Out().Map("result").Pull()
	a.Equal(11.0, i)
}

func TestBuiltin_Reduce__MixedSelection3(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := MakeOperator(core.InstanceDef{Operator: "reduce", Generics: map[string]*core.PortDef{"itemType": {Type: "number"}}})
	r.NoError(err)

	o.Out().Map("result").Bufferize()
	o.Out().Map("selection").Stream().Map("a").Bufferize()
	o.Out().Map("selection").Stream().Map("b").Bufferize()
	o.Start()

	o.In().Map("items").Push([]interface{}{1.0, 2.0})
	o.In().Map("pool").Push([]interface{}{3.0})

	a.PortPushes([]interface{}{3.0}, o.Out().Map("result"))
}
