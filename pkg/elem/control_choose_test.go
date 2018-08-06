package elem

import (
	"github.com/stretchr/testify/require"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"testing"
)

func Test_CtrlChoose__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFork := getBuiltinCfg("slang.control.Choose")
	a.NotNil(ocFork)
}

func Test_CtrlChoose__Ports(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(core.InstanceDef{Operator: "slang.control.Choose", Generics: map[string]*core.TypeDef{"itemType": {Type: "primitive"}}})
	require.NoError(t, err)

	a.Equal(core.TYPE_PRIMITIVE, o.Main().In().Map("true").Type())
	a.Equal(core.TYPE_PRIMITIVE, o.Main().In().Map("false").Type())
	a.Equal(core.TYPE_PRIMITIVE, o.Main().Out().Type())
	a.Equal(core.TYPE_PRIMITIVE, o.Delegate("chooser").Out().Map("true").Type())
	a.Equal(core.TYPE_PRIMITIVE, o.Delegate("chooser").Out().Map("false").Type())
	a.Equal(core.TYPE_BOOLEAN, o.Delegate("chooser").In().Type())
}

func Test_CtrlChoose__Works(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(core.InstanceDef{Operator: "slang.control.Choose", Generics: map[string]*core.TypeDef{"itemType": {Type: "primitive"}}})
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Delegate("chooser").Out().Bufferize()
	o.Start()

	trues := []interface{}{"Roses", 6, false, "Violets", "are", nil, 1, 2, nil, 4}
	falses := []interface{}{nil, "are", "red.", nil, nil, "blue.", "test", nil, 3, nil}
	selects := []interface{}{true, false, false, true, true, false, true, true, false, true}

	for _, v := range trues {
		o.Main().In().Map("true").Push(v)
	}
	for _, v := range falses {
		o.Main().In().Map("false").Push(v)
	}
	for i, v := range selects {
		itm := o.Delegate("chooser").Out().Pull().(map[string]interface{})
		a.Equal(trues[i], itm["true"])
		a.Equal(falses[i], itm["false"])
		o.Delegate("chooser").In().Push(v)
	}

	a.PortPushesAll([]interface{}{"Roses", "are", "red.", "Violets", "are", "blue.", 1, 2, 3, 4}, o.Main().Out())
}

func Test_CtrlChoose__ComplexItems(t *testing.T) {
	a := assertions.New(t)
	o, err := buildOperator(core.InstanceDef{
		Operator: "slang.control.Choose",
		Generics: map[string]*core.TypeDef{"itemType": {Type: "map", Map: map[string]*core.TypeDef{"red": {Type: "string"}, "blue": {Type: "string"}}}},
	})
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Delegate("chooser").Out().Bufferize()
	o.Start()

	trues := []interface{}{
		map[string]interface{}{
			"red":  "1",
			"blue": "2",
		},
		map[string]interface{}{
			"red":  "Roses",
			"blue": "Violets",
		},
		map[string]interface{}{
			"red":  "Apples",
			"blue": "Blueberries",
		},
	}
	falses := []interface{}{
		map[string]interface{}{
			"red":  "Red Bull",
			"blue": "Blues",
		},
		map[string]interface{}{
			"red":  "3",
			"blue": "4",
		},
		map[string]interface{}{
			"red":  "5",
			"blue": "6",
		},
	}
	selects := []interface{}{false, true, true}

	for _, v := range trues {
		o.Main().In().Map("true").Push(v)
	}
	for _, v := range falses {
		o.Main().In().Map("false").Push(v)
	}
	for i, v := range selects {
		itm := o.Delegate("chooser").Out().Pull().(map[string]interface{})
		a.Equal(trues[i], itm["true"])
		a.Equal(falses[i], itm["false"])
		o.Delegate("chooser").In().Push(v)
	}

	a.PortPushesAll([]interface{}{
		map[string]interface{}{"red": "Red Bull", "blue": "Blues"},
		map[string]interface{}{"red": "Roses", "blue": "Violets"},
		map[string]interface{}{"red": "Apples", "blue": "Blueberries"},
	}, o.Main().Out())
}
