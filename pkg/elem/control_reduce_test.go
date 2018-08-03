package elem

import (
	"testing"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/stretchr/testify/require"
)

func Test_CtrlReduce__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocReduce := getBuiltinCfg("slang.control.Reduce")
	a.NotNil(ocReduce)
}

func Test_CtrlReduce__NoGenerics(t *testing.T) {
	a := assertions.New(t)

	_, err := buildOperator(core.InstanceDef{Operator: "slang.control.Reduce"})
	a.Error(err)
}

func Test_CtrlReduce__InPorts(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := buildOperator(core.InstanceDef{Operator: "slang.control.Reduce", Generics: map[string]*core.TypeDef{"itemType": {Type: "number"}}, Properties: map[string]interface{}{"emptyValue": -1}})
	r.NoError(err)

	a.Equal(core.TYPE_STREAM, o.Main().In().Type())
	a.Equal(core.TYPE_STREAM, o.Delegate("selection").In().Type())

	// Item type
	itemType := core.TYPE_NUMBER
	a.Equal(itemType, o.Main().In().Stream().Type())
	a.Equal(itemType, o.Delegate("selection").In().Stream().Type())

	o, err = buildOperator(core.InstanceDef{Operator: "slang.control.Reduce", Generics: map[string]*core.TypeDef{"itemType": {Type: "string"}}, Properties: map[string]interface{}{"emptyValue": ""}})
	r.NoError(err)

	// Item type
	itemType = core.TYPE_STRING
	a.Equal(itemType, o.Main().In().Stream().Type())
	a.Equal(itemType, o.Delegate("selection").In().Stream().Type())
	r.NoError(err)
}

func Test_CtrlReduce__OutPorts(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := buildOperator(core.InstanceDef{Operator: "slang.control.Reduce", Generics: map[string]*core.TypeDef{"itemType": {Type: "number"}}, Properties: map[string]interface{}{"emptyValue": -1}})
	r.NoError(err)

	a.Equal(core.TYPE_NUMBER, o.Main().Out().Type())
	a.Equal(core.TYPE_STREAM, o.Delegate("selection").Out().Type())
	a.Equal(core.TYPE_MAP, o.Delegate("selection").Out().Stream().Type())

	// Item type
	itemType := core.TYPE_NUMBER
	a.Equal(itemType, o.Delegate("selection").Out().Stream().Map("a").Type())
	a.Equal(itemType, o.Delegate("selection").Out().Stream().Map("b").Type())

	o, err = buildOperator(core.InstanceDef{Operator: "slang.control.Reduce", Generics: map[string]*core.TypeDef{"itemType": {Type: "string"}}, Properties: map[string]interface{}{"emptyValue": ""}})
	r.NoError(err)

	// Item type
	itemType = core.TYPE_STRING
	a.Equal(itemType, o.Delegate("selection").Out().Stream().Map("a").Type())
	a.Equal(itemType, o.Delegate("selection").Out().Stream().Map("b").Type())
}

func Test_CtrlReduce__PassMarkers(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := buildOperator(core.InstanceDef{Operator: "slang.control.Reduce", Generics: map[string]*core.TypeDef{"itemType": {Type: "number"}}, Properties: map[string]interface{}{"emptyValue": -1}})
	r.NoError(err)

	o.Main().Out().Bufferize()
	o.Delegate("selection").Out().Stream().Bufferize()
	o.Start()

	bos := core.BOS{}
	eos := core.BOS{}
	o.Main().In().Stream().Push(bos)
	o.Main().In().Stream().Push(eos)
	o.Delegate("selection").In().Stream().Push(bos)
	o.Delegate("selection").In().Stream().Push(eos)

	a.PortPushesAll([]interface{}{bos, eos}, o.Main().Out())
}

func Test_CtrlReduce__SelectionFromItemsEmpty(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := buildOperator(core.InstanceDef{
		Operator: "slang.control.Reduce",
		Generics: map[string]*core.TypeDef{"itemType": {Type: "string"}},
		Properties: map[string]interface{}{"emptyValue": "empty"},
	})
	r.NoError(err)

	o.Main().Out().Bufferize()
	o.Delegate("selection").Out().Stream().Map("a").Bufferize()
	o.Delegate("selection").Out().Stream().Map("b").Bufferize()
	o.Start()

	o.Main().In().Push([]interface{}{})

	i := o.Main().Out().Pull()
	a.Equal("empty", i)
}

func Test_CtrlReduce__SelectionFromItemsSingle(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := buildOperator(core.InstanceDef{Operator: "slang.control.Reduce", Generics: map[string]*core.TypeDef{"itemType": {Type: "number"}}, Properties: map[string]interface{}{"emptyValue": -1}})
	r.NoError(err)

	o.Main().Out().Bufferize()
	o.Delegate("selection").Out().Stream().Map("a").Bufferize()
	o.Delegate("selection").Out().Stream().Map("b").Bufferize()
	o.Start()

	o.Main().In().Push([]interface{}{123.0})

	i := o.Main().Out().Pull()
	a.Equal(123.0, i)
}

func Test_CtrlReduce__SelectionFromItemsMultiple(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := buildOperator(core.InstanceDef{Operator: "slang.control.Reduce", Generics: map[string]*core.TypeDef{"itemType": {Type: "number"}}, Properties: map[string]interface{}{"emptyValue": -1}})
	r.NoError(err)

	o.Main().Out().Bufferize()
	o.Delegate("selection").Out().Stream().Map("a").Bufferize()
	o.Delegate("selection").Out().Stream().Map("b").Bufferize()
	o.Start()

	o.Main().In().Push([]interface{}{1.0, 2.0})
	o.Delegate("selection").In().Push([]interface{}{3.0})

	i := o.Delegate("selection").Out().Pull()
	a.Equal([]interface{}{map[string]interface{}{"a": 1.0, "b": 2.0}}, i)

	i = o.Main().Out().Pull()
	a.Equal(3.0, i)
}

func Test_CtrlReduce__SelectionFromPool(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := buildOperator(core.InstanceDef{Operator: "slang.control.Reduce", Generics: map[string]*core.TypeDef{"itemType": {Type: "number"}}, Properties: map[string]interface{}{"emptyValue": -1}})
	r.NoError(err)

	o.Main().Out().Bufferize()
	o.Delegate("selection").Out().Stream().Map("a").Bufferize()
	o.Delegate("selection").Out().Stream().Map("b").Bufferize()
	o.Start()

	o.Main().In().Push([]interface{}{1.0, 2.0})
	o.Delegate("selection").In().Push([]interface{}{3.0, 4.0, 5.0, 6.0})

	i := o.Delegate("selection").Out().Pull()
	a.Equal([]interface{}{map[string]interface{}{"a": 1.0, "b": 2.0}}, i)

	i = o.Delegate("selection").Out().Pull()
	a.Equal([]interface{}{
		map[string]interface{}{"a": 3.0, "b": 4.0},
		map[string]interface{}{"a": 5.0, "b": 6.0},
	}, i)
}

func Test_CtrlReduce__MixedSelection1(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := buildOperator(core.InstanceDef{Operator: "slang.control.Reduce", Generics: map[string]*core.TypeDef{"itemType": {Type: "number"}}, Properties: map[string]interface{}{"emptyValue": -1}})
	r.NoError(err)

	o.Main().Out().Bufferize()
	o.Delegate("selection").Out().Stream().Map("a").Bufferize()
	o.Delegate("selection").Out().Stream().Map("b").Bufferize()
	o.Start()

	o.Main().In().Push([]interface{}{1.0, 2.0, 3.0})
	o.Delegate("selection").In().Push([]interface{}{4.0})

	i := o.Delegate("selection").Out().Pull()
	a.Equal([]interface{}{map[string]interface{}{"a": 1.0, "b": 2.0}}, i)
	i = o.Delegate("selection").Out().Pull()
	a.Equal([]interface{}{map[string]interface{}{"a": 3.0, "b": 4.0}}, i)
}

func Test_CtrlReduce__MixedSelection2(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := buildOperator(core.InstanceDef{Operator: "slang.control.Reduce", Generics: map[string]*core.TypeDef{"itemType": {Type: "number"}}, Properties: map[string]interface{}{"emptyValue": -1}})
	r.NoError(err)

	o.Main().Out().Bufferize()
	o.Delegate("selection").Out().Stream().Map("a").Bufferize()
	o.Delegate("selection").Out().Stream().Map("b").Bufferize()
	o.Start()

	o.Main().In().Push([]interface{}{1.0, 2.0, 3.0})
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

	i = o.Main().Out().Pull()
	a.Equal(11.0, i)
}

func Test_CtrlReduce__MixedSelection3(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := buildOperator(core.InstanceDef{Operator: "slang.control.Reduce", Generics: map[string]*core.TypeDef{"itemType": {Type: "number"}}, Properties: map[string]interface{}{"emptyValue": -1}})
	r.NoError(err)

	o.Main().Out().Bufferize()
	o.Delegate("selection").Out().Stream().Map("a").Bufferize()
	o.Delegate("selection").Out().Stream().Map("b").Bufferize()
	o.Start()

	o.Main().In().Push([]interface{}{1.0, 2.0})
	o.Delegate("selection").In().Push([]interface{}{3.0})

	a.PortPushesAll([]interface{}{3.0}, o.Main().Out())
}
