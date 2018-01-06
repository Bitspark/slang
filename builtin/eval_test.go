package builtin

import (
	"slang/core"
	"slang/tests/assertions"
	"testing"

	"github.com/Knetic/govaluate"
)

func TestEvaluableExpression__Translates_Variables(t *testing.T) {
	a := assertions.New(t)

	evalExpr, _ := NewEvaluableExpression("a + 100")

	a.Equal([]string{"a"}, evalExpr.Vars())
}

func TestEvaluableExpression__Translates_AccessingFields(t *testing.T) {
	a := assertions.New(t)

	evalExpr, _ := NewEvaluableExpression("vec.x + 100")

	a.Equal([]string{"vec__x"}, evalExpr.Vars())
}

func TestEvaluableExpression__Translates_AccessingNestedFields(t *testing.T) {
	a := assertions.New(t)

	evalExpr, _ := NewEvaluableExpression(`person.address.zipcode == "ABCD"`)

	a.Equal([]string{"person__address__zipcode"}, evalExpr.Vars())
}

func TestEvaluableExpression__Evaluates_ArgumentsAcessingFields(t *testing.T) {
	a := assertions.New(t)

	evalExpr, _ := NewEvaluableExpression("(vec.0.x + vec.1.x) * (vec.0.y + vec.1.y)")
	params := NewFlatMapParameters(map[string]interface{}{
		"vec": []interface{}{
			map[string]interface{}{
				"x": 1,
				"y": 2,
			},
			map[string]interface{}{
				"x": 3,
				"y": 4,
			},
		},
	})
	result, _ := evalExpr.Evaluate(params)
	a.Equal(24.0, result)
}

func TestEvaluableExpression__Translates_Combined(t *testing.T) {
	a := assertions.New(t)

	evalExpr, _ := NewEvaluableExpression(`a * vec.x + vec.y`)

	a.Equal([]string{"a", "vec__x", "vec__y"}, evalExpr.Vars())
}

func TestFlatMapParameters__NestingLevel0(t *testing.T) {
	a := assertions.New(t)

	params := NewFlatMapParameters(map[string]interface{}{
		"foo": 100,
		"bar": 200,
	})

	a.Equal(govaluate.MapParameters(map[string]interface{}{
		"foo": 100,
		"bar": 200,
	}), params)
}

func TestFlatMapParameters__NestingLevel0_Arrays(t *testing.T) {
	a := assertions.New(t)

	params := NewFlatMapParameters(map[string]interface{}{
		"foo": 100,
		"bar": 200,
		"l":   []interface{}{1, 2},
	})

	a.Equal(govaluate.MapParameters(map[string]interface{}{
		"foo":  100,
		"bar":  200,
		"l__0": 1,
		"l__1": 2,
	}), params)
}

func TestFlatMapParameters__NestingLevel1(t *testing.T) {
	a := assertions.New(t)

	params := NewFlatMapParameters(map[string]interface{}{
		"foo": 100,
		"bar": 200,
		"vec": map[string]interface{}{
			"x": 1,
			"y": 2,
		},
	})

	a.Equal(govaluate.MapParameters(map[string]interface{}{
		"foo":    100,
		"bar":    200,
		"vec__x": 1,
		"vec__y": 2,
	}), params)
}

func TestFlatMapParameters__NestingLevel2(t *testing.T) {
	a := assertions.New(t)

	params := NewFlatMapParameters(map[string]interface{}{
		"foo": 100,
		"bar": 200,
		"vec": map[string]interface{}{
			"x": 1,
			"y": 2,
		},
		"person": map[string]interface{}{
			"phone": map[string]interface{}{
				"mobile": "0123/4567890",
			},
			"name": map[string]interface{}{
				"first": "Paul J.",
				"last":  "Morrison",
			},
		},
	})

	a.Equal(govaluate.MapParameters(map[string]interface{}{
		"foo":                   100,
		"bar":                   200,
		"vec__x":                1,
		"vec__y":                2,
		"person__phone__mobile": "0123/4567890",
		"person__name__first":   "Paul J.",
		"person__name__last":    "Morrison",
	}), params)
}

func TestFlatMapParameters__NestingLevel3(t *testing.T) {
	a := assertions.New(t)

	params := NewFlatMapParameters(map[string]interface{}{
		"foo": 100,
		"bar": 200,
		"vec": map[string]interface{}{
			"x": 1,
			"y": 2,
		},
		"root": map[string]interface{}{
			"left": map[string]interface{}{
				"left": map[string]interface{}{
					"i": 100,
				},
				"right": map[string]interface{}{

					"i": 101,
				},
			},
			"right": map[string]interface{}{
				"left": map[string]interface{}{
					"i": 110,
				},
				"right": map[string]interface{}{

					"i": 111,
				},
			},
		},
	})

	a.Equal(govaluate.MapParameters(map[string]interface{}{
		"foo":                   100,
		"bar":                   200,
		"vec__x":                1,
		"vec__y":                2,
		"root__left__left__i":   100,
		"root__left__right__i":  101,
		"root__right__left__i":  110,
		"root__right__right__i": 111,
	}), params)
}

func TestFlatMapParameters__ComplexMixed(t *testing.T) {
	a := assertions.New(t)

	params := NewFlatMapParameters(map[string]interface{}{
		"foo": 100,
		"bar": 200,
		"vec": []interface{}{
			map[string]interface{}{
				"x": 1,
				"y": 2,
			},
			map[string]interface{}{
				"x": 3,
				"y": 4,
			},
		},
		"person": []interface{}{
			map[string]interface{}{
				"phone": map[string]interface{}{
					"mobile": "0123/4567890",
				},
				"name": map[string]interface{}{
					"first": "Taleh",
					"last":  "Didover",
				},
			},
			map[string]interface{}{
				"phone": map[string]interface{}{
					"mobile": "0123/4567890",
				},
				"name": map[string]interface{}{
					"first": "Julian",
					"last":  "Matschinske",
				},
			},
		},
	})

	a.Equal(govaluate.MapParameters(map[string]interface{}{
		"foo":                      100,
		"bar":                      200,
		"vec__0__x":                1,
		"vec__0__y":                2,
		"vec__1__x":                3,
		"vec__1__y":                4,
		"person__0__phone__mobile": "0123/4567890",
		"person__0__name__first":   "Taleh",
		"person__0__name__last":    "Didover",
		"person__1__phone__mobile": "0123/4567890",
		"person__1__name__first":   "Julian",
		"person__1__name__last":    "Matschinske",
	}), params)
}

func TestBuiltin_Eval__CreatorFuncIsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFunction := getCreatorFunc("eval")
	a.NotNil(ocFunction)
}

func TestBuiltin_Eval__NilProperties(t *testing.T) {
	fo, err := createOpEval(core.InstanceDef{Operator: "eval"}, nil)

	if fo != nil || err == nil {
		t.Error("expected error")
	}
}

func TestBuiltin_Eval__EmptyExpression(t *testing.T) {
	fo, err := createOpEval(core.InstanceDef{Operator: "eval", Properties: map[string]interface{}{"expression": ""}}, nil)

	if fo != nil || err == nil {
		t.Error("expected error")
	}
}

func TestBuiltin_Eval__InvalidExpression(t *testing.T) {
	fo, err := createOpEval(core.InstanceDef{Operator: "eval", Properties: map[string]interface{}{"expression": "+"}}, nil)

	if fo != nil || err == nil {
		t.Error("expected error")
	}
}

func TestBuiltin_Eval__Add(t *testing.T) {
	fo, err := createOpEval(core.InstanceDef{Operator: "eval", Properties: map[string]interface{}{"expression": "a+b"}}, nil)

	if fo == nil {
		t.Error("operator not defined")
	}

	if err != nil {
		t.Error(err.Error())
	}

	if fo.In().Type() != core.TYPE_MAP {
		t.Error("expected map")
	}

	if fo.In().Map("a").Type() != core.TYPE_ANY {
		t.Error("expected any input")
	}

	if fo.In().Map("b").Type() != core.TYPE_ANY {
		t.Error("expected any input")
	}

	fo.Out().Bufferize()

	go fo.Start()

	fo.In().Push(map[string]interface{}{"a": 1.0, "b": 2.0})
	fo.In().Push(map[string]interface{}{"a": -5.0, "b": 2.5})
	fo.In().Push(map[string]interface{}{"a": 0.0, "b": 333.0})

	if fo.Out().Pull() != 3.0 {
		t.Error("wrong output")
	}

	if fo.Out().Pull() != -2.5 {
		t.Error("wrong output")
	}

	if fo.Out().Pull() != 333.0 {
		t.Error("wrong output")
	}
}

func TestBuiltin_Eval__BoolArith(t *testing.T) {
	fo, err := createOpEval(core.InstanceDef{Operator: "eval", Properties: map[string]interface{}{"expression": "a && (b != c)"}}, nil)

	if fo == nil {
		t.Error("operator not defined")
	}

	if err != nil {
		t.Error(err.Error())
	}

	if fo.In().Type() != core.TYPE_MAP {
		t.Error("expected map")
	}

	if fo.In().Map("a").Type() != core.TYPE_ANY {
		t.Error("expected any input")
	}

	if fo.In().Map("b").Type() != core.TYPE_ANY {
		t.Error("expected any input")
	}

	if fo.In().Map("c").Type() != core.TYPE_ANY {
		t.Error("expected any input")
	}

	fo.Out().Bufferize()

	go fo.Start()

	fo.In().Push(map[string]interface{}{"a": true, "b": true, "c": false})
	fo.In().Push(map[string]interface{}{"a": false, "b": false, "c": false})
	fo.In().Push(map[string]interface{}{"a": false, "b": false, "c": true})
	fo.In().Push(map[string]interface{}{"a": true, "b": false, "c": true})
	fo.In().Push(map[string]interface{}{"a": true, "b": false, "c": false})

	if fo.Out().Pull() != true {
		t.Error("wrong output")
	}

	if fo.Out().Pull() != false {
		t.Error("wrong output")
	}

	if fo.Out().Pull() != false {
		t.Error("wrong output")
	}

	if fo.Out().Pull() != true {
		t.Error("wrong output")
	}

	if fo.Out().Pull() != false {
		t.Error("wrong output")
	}
}
