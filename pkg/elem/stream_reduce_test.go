package elem

import (
	"testing"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_CtrlReduce__IsRegistered(t *testing.T) {
	Init()
	a := assertions.New(t)

	ocReduce := getBuiltinCfg(uuid.MustParse("b95e6da8-9770-4a04-a73d-cdfe2081870f"))
	a.NotNil(ocReduce)
}

func Test_CtrlReduce__NoGenerics(t *testing.T) {
	Init()
	a := assertions.New(t)

	_, err := buildOperator(core.InstanceDef{Operator: streamReduceCfg.blueprint.Id})
	a.Error(err)
}

func Test_CtrlReduce__InPorts(t *testing.T) {
	Init()
	a := assertions.New(t)
	r := require.New(t)

	o, err := buildOperator(core.InstanceDef{Operator: streamReduceCfg.blueprint.Id, Generics: map[string]*core.TypeDef{"itemType": {Type: "number"}}, Properties: map[string]interface{}{"emptyValue": -1}})
	r.NoError(err)

	// Item type
	itemType := core.TYPE_NUMBER
	a.Equal(itemType, o.Main().In().Stream().Type())
	a.Equal(itemType, o.Delegate("reducer").In().Type())

	o, err = buildOperator(core.InstanceDef{Operator: streamReduceCfg.blueprint.Id, Generics: map[string]*core.TypeDef{"itemType": {Type: "string"}}, Properties: map[string]interface{}{"emptyValue": ""}})
	r.NoError(err)

	// Item type
	itemType = core.TYPE_STRING
	a.Equal(itemType, o.Main().In().Stream().Type())
	a.Equal(itemType, o.Delegate("reducer").In().Type())
	r.NoError(err)
}

func Test_CtrlReduce__OutPorts(t *testing.T) {
	Init()
	a := assertions.New(t)
	r := require.New(t)

	o, err := buildOperator(core.InstanceDef{Operator: streamReduceCfg.blueprint.Id, Generics: map[string]*core.TypeDef{"itemType": {Type: "number"}}, Properties: map[string]interface{}{"emptyValue": -1}})
	r.NoError(err)

	a.Equal(core.TYPE_NUMBER, o.Main().Out().Type())
	a.Equal(core.TYPE_MAP, o.Delegate("reducer").Out().Type())

	// Item type
	itemType := core.TYPE_NUMBER
	a.Equal(itemType, o.Delegate("reducer").Out().Map("a").Type())
	a.Equal(itemType, o.Delegate("reducer").Out().Map("b").Type())

	o, err = buildOperator(core.InstanceDef{Operator: streamReduceCfg.blueprint.Id, Generics: map[string]*core.TypeDef{"itemType": {Type: "string"}}, Properties: map[string]interface{}{"emptyValue": ""}})
	r.NoError(err)

	// Item type
	itemType = core.TYPE_STRING
	a.Equal(itemType, o.Delegate("reducer").Out().Map("a").Type())
	a.Equal(itemType, o.Delegate("reducer").Out().Map("b").Type())
}

func Test_CtrlReduce__PassMarkers(t *testing.T) {
	Init()
	a := assertions.New(t)
	r := require.New(t)

	o, err := buildOperator(core.InstanceDef{Operator: streamReduceCfg.blueprint.Id, Generics: map[string]*core.TypeDef{"itemType": {Type: "number"}}, Properties: map[string]interface{}{"emptyValue": -1}})
	r.NoError(err)

	o.Main().Out().Bufferize()
	o.Delegate("reducer").Out().Bufferize()
	o.Start()

	bos := core.BOS{}
	eos := core.BOS{}
	o.Main().In().Stream().Push(bos)
	o.Main().In().Stream().Push(eos)

	a.PortPushesAll([]interface{}{bos, eos}, o.Main().Out())
}

func Test_CtrlReduce__SelectionFromItemsEmpty(t *testing.T) {
	Init()
	a := assertions.New(t)
	r := require.New(t)

	o, err := buildOperator(core.InstanceDef{
		Operator:   streamReduceCfg.blueprint.Id,
		Generics:   map[string]*core.TypeDef{"itemType": {Type: "string"}},
		Properties: map[string]interface{}{"emptyValue": "empty"},
	})
	r.NoError(err)

	o.Main().Out().Bufferize()
	o.Delegate("reducer").Out().Bufferize()
	o.Start()

	o.Main().In().Push([]interface{}{})

	i := o.Main().Out().Pull()
	a.Equal("empty", i)
}

func Test_CtrlReduce__SelectionFromItemsSingle(t *testing.T) {
	Init()
	a := assertions.New(t)
	r := require.New(t)

	o, err := buildOperator(core.InstanceDef{Operator: streamReduceCfg.blueprint.Id, Generics: map[string]*core.TypeDef{"itemType": {Type: "number"}}, Properties: map[string]interface{}{"emptyValue": -1}})
	r.NoError(err)

	o.Main().Out().Bufferize()
	o.Delegate("reducer").Out().Bufferize()
	o.Start()

	o.Main().In().Push([]interface{}{123.0})

	i := o.Main().Out().Pull()
	a.Equal(123.0, i)
}

func Test_CtrlReduce__SelectionFromItemsMultiple(t *testing.T) {
	Init()
	a := assertions.New(t)
	r := require.New(t)

	o, err := buildOperator(core.InstanceDef{Operator: streamReduceCfg.blueprint.Id, Generics: map[string]*core.TypeDef{"itemType": {Type: "number"}}, Properties: map[string]interface{}{"emptyValue": -1}})
	r.NoError(err)

	o.Main().Out().Bufferize()
	o.Delegate("reducer").Out().Bufferize()
	o.Start()

	o.Main().In().Push([]interface{}{1.0, 2.0})
	o.Delegate("reducer").In().Push(3.0)

	i := o.Delegate("reducer").Out().Pull()
	a.Equal(map[string]interface{}{"a": 1.0, "b": 2.0}, i)

	i = o.Main().Out().Pull()
	a.Equal(3.0, i)
}

func Test_CtrlReduce__SelectionFromPool(t *testing.T) {
	Init()
	a := assertions.New(t)
	r := require.New(t)

	o, err := buildOperator(core.InstanceDef{
		Operator: streamReduceCfg.blueprint.Id,
		Generics: map[string]*core.TypeDef{"itemType": {Type: "number"}},
		Properties: map[string]interface{}{"emptyValue": -1},
	})
	r.NoError(err)

	o.Main().Out().Bufferize()
	o.Delegate("reducer").Out().Bufferize()
	o.Start()

	o.Main().In().Push([]interface{}{1.0, 1.0, 1.0, 1.0})
	o.Delegate("reducer").In().Push(2.0)
	o.Delegate("reducer").In().Push(2.0)
	o.Delegate("reducer").In().Push(4.0)

	i := o.Delegate("reducer").Out().Pull()
	a.Equal(map[string]interface{}{"a": 1.0, "b": 1.0}, i)
	i = o.Delegate("reducer").Out().Pull()
	a.Equal(map[string]interface{}{"a": 1.0, "b": 1.0}, i)
	i = o.Delegate("reducer").Out().Pull()
	a.Equal(map[string]interface{}{"a": 2.0, "b": 2.0}, i)
}
