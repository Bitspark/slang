package elem

import (
	"github.com/stretchr/testify/require"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"testing"
)

func Test_CtrlSwitch__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocSwitch := getBuiltinCfg("slang.control.Switch")
	a.NotNil(ocSwitch)
}

func Test_CtrlSwitch__Ports(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.control.Switch",
			Generics: map[string]*core.TypeDef{
				"inType": {Type: "string"},
				"outType": {Type: "number"},
				"selectType": {Type: "boolean"},
			},
			Properties: map[string]interface{}{
				"cases": []interface{}{},
			},
		},
	)
	r.NoError(err)
	r.NotNil(o)

	a.Equal(core.TYPE_STRING, o.Main().In().Map("item").Type())
	a.Equal(core.TYPE_BOOLEAN, o.Main().In().Map("select").Type())
	a.Equal(core.TYPE_NUMBER, o.Main().Out().Type())
}

func Test_CtrlSwitch__Delegates_Bool(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.control.Switch",
			Generics: map[string]*core.TypeDef{
				"inType": {Type: "string"},
				"outType": {Type: "number"},
				"selectType": {Type: "boolean"},
			},
			Properties: map[string]interface{}{
				"cases": []interface{}{true, false},
			},
		},
	)
	r.NoError(err)
	r.NotNil(o)

	a.NotNil(o.Delegate("true"))
	a.NotNil(o.Delegate("false"))
}

func Test_CtrlSwitch__Delegates_String(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.control.Switch",
			Generics: map[string]*core.TypeDef{
				"inType": {Type: "string"},
				"outType": {Type: "number"},
				"selectType": {Type: "string"},
			},
			Properties: map[string]interface{}{
				"cases": []interface{}{"test1", "test2", "test3"},
			},
		},
	)
	r.NoError(err)
	r.NotNil(o)

	a.NotNil(o.Delegate("test1"))
	a.NotNil(o.Delegate("test2"))
	a.NotNil(o.Delegate("test3"))
}

func Test_CtrlSwitch__Redirect_Number(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.control.Switch",
			Generics: map[string]*core.TypeDef{
				"inType": {Type: "string"},
				"outType": {Type: "number"},
				"selectType": {Type: "number"},
			},
			Properties: map[string]interface{}{
				"cases": []interface{}{1, 5, 12},
			},
		},
	)
	r.NoError(err)
	r.NotNil(o)

	a.NotNil(o.Delegate("1"))
	a.NotNil(o.Delegate("5"))
	a.NotNil(o.Delegate("12"))

	o.Main().Out().Bufferize()
	o.Delegate("1").Out().Bufferize()
	o.Delegate("5").Out().Bufferize()
	o.Delegate("12").Out().Bufferize()

	o.Start()

	o.Main().In().Map("item").Push("hallo")
	o.Main().In().Map("item").Push("slang")
	o.Main().In().Map("item").Push(":)")
	o.Main().In().Map("select").Push(5)
	o.Main().In().Map("select").Push(1)
	o.Main().In().Map("select").Push(12)

	a.Equal("hallo", o.Delegate("5").Out().Pull())
	o.Delegate("5").In().Push(11)
	a.Equal("slang", o.Delegate("1").Out().Pull())
	o.Delegate("1").In().Push(12)
	a.Equal(":)", o.Delegate("12").Out().Pull())
	o.Delegate("12").In().Push(13)

	a.Equal(11, o.Main().Out().Pull())
	a.Equal(12, o.Main().Out().Pull())
	a.Equal(13, o.Main().Out().Pull())
}

