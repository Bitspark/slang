package elem

import (
	"testing"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
)

func Test_DataEvaluate__TranslatesVariables(t *testing.T) {
	a := assertions.New(t)

	evalExpr, _ := newEvaluableExpression("a + 100")

	a.Equal([]string{"a"}, evalExpr.Vars())
}

func Test_DataEvaluate__IsRegistered(t *testing.T) {
	a := assertions.New(t)
	a.True(IsRegistered(dataEvaluateId))
}

func Test_DataEvaluate__NilProperties(t *testing.T) {
	a := assertions.New(t)
	_, err := buildOperator(core.InstanceDef{Operator: dataEvaluateId})
	a.Error(err)
}

func Test_DataEvaluate__EmptyExpression(t *testing.T) {
	a := assertions.New(t)
	_, err := buildOperator(core.InstanceDef{Operator: dataEvaluateId, Properties: map[string]interface{}{"expression": ""}})
	a.Error(err)
}

func Test_DataEvaluate__InvalidExpression(t *testing.T) {
	a := assertions.New(t)
	_, err := buildOperator(core.InstanceDef{Operator: dataEvaluateId, Properties: map[string]interface{}{"expression": "+"}})
	a.Error(err)
}

func Test_DataEvaluate__Add(t *testing.T) {
	a := assertions.New(t)
	fo, err := buildOperator(core.InstanceDef{
		Operator: dataEvaluateId,
		Properties: map[string]interface{}{
			"expression": "a+b",
			"variables":  []interface{}{"a", "b"},
		},
	})
	a.NoError(err)
	a.NotNil(fo)
	fo.Main().Out().Bufferize()

	go fo.Start()

	fo.Main().In().Push(map[string]interface{}{"a": 1.0, "b": 2.0})
	fo.Main().In().Push(map[string]interface{}{"a": -5.0, "b": 2.5})
	fo.Main().In().Push(map[string]interface{}{"a": 0.0, "b": 333.0})

	a.PortPushesAll([]interface{}{3.0, -2.5, 333.0}, fo.Main().Out())
}

func Test_DataEvaluate__Floor(t *testing.T) {
	a := assertions.New(t)
	fo, err := buildOperator(core.InstanceDef{
		Operator: dataEvaluateId,
		Properties: map[string]interface{}{
			"expression": "floor(a)",
			"variables":  []interface{}{"a"},
		},
	})
	a.NoError(err)
	a.NotNil(fo)
	fo.Main().Out().Bufferize()

	go fo.Start()

	fo.Main().In().Push(map[string]interface{}{"a": 1.0})
	fo.Main().In().Push(map[string]interface{}{"a": 1.1})
	fo.Main().In().Push(map[string]interface{}{"a": 2.9})

	a.PortPushesAll([]interface{}{1.0, 1.0, 2.0}, fo.Main().Out())
}

func Test_DataEvaluate__Ceil(t *testing.T) {
	a := assertions.New(t)
	fo, err := buildOperator(core.InstanceDef{
		Operator: dataEvaluateId,
		Properties: map[string]interface{}{
			"expression": "ceil(a)",
			"variables":  []interface{}{"a"},
		},
	})
	a.NoError(err)
	a.NotNil(fo)
	fo.Main().Out().Bufferize()

	go fo.Start()

	fo.Main().In().Push(map[string]interface{}{"a": 1.0})
	fo.Main().In().Push(map[string]interface{}{"a": 1.1})
	fo.Main().In().Push(map[string]interface{}{"a": 2.9})

	a.PortPushesAll([]interface{}{1.0, 2.0, 3.0}, fo.Main().Out())
}

func Test_DataEvaluate__IsNull(t *testing.T) {
	a := assertions.New(t)
	fo, err := buildOperator(core.InstanceDef{
		Operator: dataEvaluateId,
		Properties: map[string]interface{}{
			"expression": "isNull(a)",
			"variables":  []interface{}{"a"},
		},
	})
	a.NoError(err)
	a.NotNil(fo)
	fo.Main().Out().Bufferize()

	go fo.Start()

	fo.Main().In().Push(map[string]interface{}{"a": 1.0})
	fo.Main().In().Push(map[string]interface{}{"a": nil})
	fo.Main().In().Push(map[string]interface{}{"a": "testtest"})

	a.PortPushesAll([]interface{}{false, true, false}, fo.Main().Out())
}

func Test_DataEvaluate__Pow(t *testing.T) {
	a := assertions.New(t)
	fo, err := buildOperator(core.InstanceDef{
		Operator: dataEvaluateId,
		Properties: map[string]interface{}{
			"expression": "pow(a, b)",
			"variables":  []interface{}{"a", "b"},
		},
	})
	a.NoError(err)
	a.NotNil(fo)
	fo.Main().Out().Bufferize()

	go fo.Start()

	fo.Main().In().Push(map[string]interface{}{"a": 2.0, "b": 0.0})
	fo.Main().In().Push(map[string]interface{}{"a": 1.1, "b": 1.0})
	fo.Main().In().Push(map[string]interface{}{"a": 9.0, "b": 0.5})
	fo.Main().In().Push(map[string]interface{}{"a": 2.0, "b": 3.0})

	a.PortPushesAll([]interface{}{1.0, 1.1, 3.0, 8.0}, fo.Main().Out())
}

func Test_DataEvaluate__BoolArith(t *testing.T) {
	a := assertions.New(t)
	fo, err := buildOperator(core.InstanceDef{
		Operator: dataEvaluateId,
		Properties: map[string]interface{}{
			"expression": "a && (b != c)",
			"variables":  []interface{}{"a", "b", "c"},
		},
	})
	a.NoError(err)
	a.NotNil(fo)
	fo.Main().Out().Bufferize()

	go fo.Start()

	fo.Main().In().Push(map[string]interface{}{"a": true, "b": 1.0, "c": 2.0})
	fo.Main().In().Push(map[string]interface{}{"a": false, "b": 8.0, "c": 8.0})
	fo.Main().In().Push(map[string]interface{}{"a": false, "b": 3.0, "c": 2.0})
	fo.Main().In().Push(map[string]interface{}{"a": true, "b": 1.0, "c": 0.0})
	fo.Main().In().Push(map[string]interface{}{"a": true, "b": 8.0, "c": 8.0})

	a.PortPushesAll([]interface{}{true, false, false, true, false}, fo.Main().Out())
}
