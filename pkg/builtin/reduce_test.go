package builtin

import (
	"testing"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/stretchr/testify/require"
)

func TestBuiltin_Reduce__CreatorFuncIsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocReduce := getBuiltinCfg("slang.reduce")
	a.NotNil(ocReduce)
}

func TestBuiltin_Reduce__NoGenerics(t *testing.T) {
	a := assertions.New(t)

	_, err := MakeOperator(core.InstanceDef{Operator: "slang.reduce"})
	a.Error(err)
}

func TestBuiltin_Reduce__InPorts(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := MakeOperator(core.InstanceDef{Operator: "slang.reduce", Generics: map[string]*core.PortDef{"itemType": {Type: "number"}}})
	r.NoError(err)

	a.Equal(core.TYPE_STREAM, o.In().Type())
	a.Equal(core.TYPE_STREAM, o.Delegate("selection").In().Type())

	// Item type
	itemType := core.TYPE_NUMBER
	a.Equal(itemType, o.In().Stream().Type())
	a.Equal(itemType, o.Delegate("selection").In().Stream().Type())

	o, err = MakeOperator(core.InstanceDef{Operator: "slang.reduce", Generics: map[string]*core.PortDef{"itemType": {Type: "string"}}})
	r.NoError(err)

	// Item type
	itemType = core.TYPE_STRING
	a.Equal(itemType, o.In().Stream().Type())
	a.Equal(itemType, o.Delegate("selection").In().Stream().Type())
	r.NoError(err)
}

func TestBuiltin_Reduce__OutPorts(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := MakeOperator(core.InstanceDef{Operator: "slang.reduce", Generics: map[string]*core.PortDef{"itemType": {Type: "number"}}})
	r.NoError(err)

	a.Equal(core.TYPE_NUMBER, o.Out().Type())
	a.Equal(core.TYPE_STREAM, o.Delegate("selection").Out().Type())
	a.Equal(core.TYPE_MAP, o.Delegate("selection").Out().Stream().Type())

	// Item type
	itemType := core.TYPE_NUMBER
	a.Equal(itemType, o.Delegate("selection").Out().Stream().Map("a").Type())
	a.Equal(itemType, o.Delegate("selection").Out().Stream().Map("b").Type())

	o, err = MakeOperator(core.InstanceDef{Operator: "slang.reduce", Generics: map[string]*core.PortDef{"itemType": {Type: "string"}}})
	r.NoError(err)

	// Item type
	itemType = core.TYPE_STRING
	a.Equal(itemType, o.Delegate("selection").Out().Stream().Map("a").Type())
	a.Equal(itemType, o.Delegate("selection").Out().Stream().Map("b").Type())
}

func TestBuiltin_Reduce__PassMarkers(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := MakeOperator(core.InstanceDef{Operator: "slang.reduce", Generics: map[string]*core.PortDef{"itemType": {Type: "number"}}})
	r.NoError(err)

	o.Out().Bufferize()
	o.Delegate("selection").Out().Stream().Bufferize()
	o.Start()

	bos := core.BOS{}
	eos := core.BOS{}
	o.In().Stream().Push(bos)
	o.In().Stream().Push(eos)
	o.Delegate("selection").In().Stream().Push(bos)
	o.Delegate("selection").In().Stream().Push(eos)

	a.PortPushes([]interface{}{bos, eos}, o.Out())
}

func TestBuiltin_Reduce__SelectionFromItemsEmpty(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := MakeOperator(core.InstanceDef{
		Operator: "slang.reduce",
		Generics: map[string]*core.PortDef{"itemType": {Type: "string"}},
		Properties: map[string]interface{}{"emptyValue": "empty"},
	})
	r.NoError(err)

	o.Out().Bufferize()
	o.Delegate("selection").Out().Stream().Map("a").Bufferize()
	o.Delegate("selection").Out().Stream().Map("b").Bufferize()
	o.Start()

	o.In().Push([]interface{}{})

	i := o.Out().Pull()
	a.Equal("empty", i)
}

func TestBuiltin_Reduce__SelectionFromItemsSingle(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := MakeOperator(core.InstanceDef{Operator: "slang.reduce", Generics: map[string]*core.PortDef{"itemType": {Type: "number"}}})
	r.NoError(err)

	o.Out().Bufferize()
	o.Delegate("selection").Out().Stream().Map("a").Bufferize()
	o.Delegate("selection").Out().Stream().Map("b").Bufferize()
	o.Start()

	o.In().Push([]interface{}{123.0})

	i := o.Out().Pull()
	a.Equal(123.0, i)
}

func TestBuiltin_Reduce__SelectionFromItemsMultiple(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := MakeOperator(core.InstanceDef{Operator: "slang.reduce", Generics: map[string]*core.PortDef{"itemType": {Type: "number"}}})
	r.NoError(err)

	o.Out().Bufferize()
	o.Delegate("selection").Out().Stream().Map("a").Bufferize()
	o.Delegate("selection").Out().Stream().Map("b").Bufferize()
	o.Start()

	o.In().Push([]interface{}{1.0, 2.0})
	o.Delegate("selection").In().Push([]interface{}{3.0})

	i := o.Delegate("selection").Out().Pull()
	a.Equal([]interface{}{map[string]interface{}{"a": 1.0, "b": 2.0}}, i)

	i = o.Out().Pull()
	a.Equal(3.0, i)
}

func TestBuiltin_Reduce__SelectionFromPool(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := MakeOperator(core.InstanceDef{Operator: "slang.reduce", Generics: map[string]*core.PortDef{"itemType": {Type: "number"}}})
	r.NoError(err)

	o.Out().Bufferize()
	o.Delegate("selection").Out().Stream().Map("a").Bufferize()
	o.Delegate("selection").Out().Stream().Map("b").Bufferize()
	o.Start()

	o.In().Push([]interface{}{1.0, 2.0})
	o.Delegate("selection").In().Push([]interface{}{3.0, 4.0, 5.0, 6.0})

	i := o.Delegate("selection").Out().Pull()
	a.Equal([]interface{}{map[string]interface{}{"a": 1.0, "b": 2.0}}, i)

	i = o.Delegate("selection").Out().Pull()
	a.Equal([]interface{}{
		map[string]interface{}{"a": 3.0, "b": 4.0},
		map[string]interface{}{"a": 5.0, "b": 6.0},
	}, i)
}

func TestBuiltin_Reduce__MixedSelection1(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := MakeOperator(core.InstanceDef{Operator: "slang.reduce", Generics: map[string]*core.PortDef{"itemType": {Type: "number"}}})
	r.NoError(err)

	o.Out().Bufferize()
	o.Delegate("selection").Out().Stream().Map("a").Bufferize()
	o.Delegate("selection").Out().Stream().Map("b").Bufferize()
	o.Start()

	o.In().Push([]interface{}{1.0, 2.0, 3.0})
	o.Delegate("selection").In().Push([]interface{}{4.0})

	i := o.Delegate("selection").Out().Pull()
	a.Equal([]interface{}{map[string]interface{}{"a": 1.0, "b": 2.0}}, i)
	i = o.Delegate("selection").Out().Pull()
	a.Equal([]interface{}{map[string]interface{}{"a": 3.0, "b": 4.0}}, i)
}

func TestBuiltin_Reduce__MixedSelection2(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := MakeOperator(core.InstanceDef{Operator: "slang.reduce", Generics: map[string]*core.PortDef{"itemType": {Type: "number"}}})
	r.NoError(err)

	o.Out().Bufferize()
	o.Delegate("selection").Out().Stream().Map("a").Bufferize()
	o.Delegate("selection").Out().Stream().Map("b").Bufferize()
	o.Start()

	o.In().Push([]interface{}{1.0, 2.0, 3.0})
	o.Delegate("selection").In().Push([]interface{}{4.0, 5.0, 6.0})
	o.Delegate("selection").In().Push([]interface{}{7.0, 8.0, 9.0})
	o.Delegate("selection").In().Push([]interface{}{10.0})
	o.Delegate("selection").In().Push([]interface{}{11.0})

	i := o.Delegate("selection").Out().Pull()
	a.Equal([]interface{}{map[string]interface{}{"a": 1.0, "b": 2.0}}, i)

	i = o.Delegate("selection").Out().Pull()
	a.Equal([]interface{}{
		map[string]interface{}{"a": 3.0, "b": 4.0},
		map[string]interface{}{"a": 5.0, "b": 6.0},
	}, i)

	i = o.Delegate("selection").Out().Pull()
	a.Equal([]interface{}{
		map[string]interface{}{"a": 7.0, "b": 8.0},
	}, i)

	i = o.Delegate("selection").Out().Pull()
	a.Equal([]interface{}{
		map[string]interface{}{"a": 9.0, "b": 10.0},
	}, i)

	i = o.Out().Pull()
	a.Equal(11.0, i)
}

func TestBuiltin_Reduce__MixedSelection3(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := MakeOperator(core.InstanceDef{Operator: "slang.reduce", Generics: map[string]*core.PortDef{"itemType": {Type: "number"}}})
	r.NoError(err)

	o.Out().Bufferize()
	o.Delegate("selection").Out().Stream().Map("a").Bufferize()
	o.Delegate("selection").Out().Stream().Map("b").Bufferize()
	o.Start()

	o.In().Push([]interface{}{1.0, 2.0})
	o.Delegate("selection").In().Push([]interface{}{3.0})

	a.PortPushes([]interface{}{3.0}, o.Out())
}