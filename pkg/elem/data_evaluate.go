package elem

import (
	"errors"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Knetic/govaluate"
	"math"
)

type EvaluableExpression struct {
	govaluate.EvaluableExpression
}

func floatify(val interface{}) (float64, bool) {
	if f, ok := val.(float64); ok {
		return f, true
	}
	if i, ok := val.(int); ok {
		return float64(i), true
	}
	return 0.0, false
}

type mathFunc1 func(float64) float64
type mathFunc2 func(float64, float64) float64

func makeMathFunc1(f mathFunc1, args ...interface{}) govaluate.ExpressionFunction {
	return func(args ...interface{}) (interface{}, error) {
		if fval, ok := floatify(args[0]); ok {
			return f(fval), nil
		}
		return nil, errors.New("wrong type")
	}
}

func makeMathFunc2(f mathFunc2, args ...interface{}) govaluate.ExpressionFunction {
	return func(args ...interface{}) (interface{}, error) {
		fval1, ok1 := floatify(args[0])
		fval2, ok2 := floatify(args[1])
		if ok1 && ok2 {
			return f(fval1, fval2), nil
		}
		return nil, errors.New("wrong type")
	}
}

func evalFunctions() map[string]govaluate.ExpressionFunction {
	functions := map[string]govaluate.ExpressionFunction {
		"floor": makeMathFunc1(math.Floor),
		"ceil": makeMathFunc1(math.Ceil),
		"sqrt": makeMathFunc1(math.Sqrt),
		"sin": makeMathFunc1(math.Sin),
		"asin": makeMathFunc1(math.Asin),
		"cos": makeMathFunc1(math.Cos),
		"acos": makeMathFunc1(math.Acos),
		"tan": makeMathFunc1(math.Tan),
		"atan": makeMathFunc1(math.Atan),
		"pow": makeMathFunc2(math.Pow),
		"atan2": makeMathFunc2(math.Atan2),
		"isNull": func(args ...interface{}) (interface{}, error) {
			if len(args) == 0 {
				return true, nil
			}
			return args[0] == nil, nil
		},
		"len": func(args ...interface{}) (interface{}, error) {
			if len(args) == 0 {
				return true, nil
			}
			return float64(len(args[0].(string))), nil
		},
	}
	return functions
}

func newEvaluableExpression(expression string) (*EvaluableExpression, error) {
	goEvalExpr, err := govaluate.NewEvaluableExpressionWithFunctions(expression, evalFunctions())
	if err == nil {
		return &EvaluableExpression{*goEvalExpr}, nil
	}
	return nil, err
}

var dataEvaluateCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type:    "map",
					Map: map[string]*core.TypeDef{
						"{variables}": {
							Type: "primitive",
						},
					},
				},
				Out: core.TypeDef{
					Type: "primitive",
				},
			},
		},
		PropertyDefs: map[string]*core.TypeDef{
			"expression": {
				Type: "string",
			},
			"variables": {
				Type: "stream",
				Stream: &core.TypeDef{
					Type: "string",
				},
			},
		},
	},
	opFunc: func(op *core.Operator) {
		expr, _ := newEvaluableExpression(op.Property("expression").(string))
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			i := in.Pull()

			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			if m, ok := i.(map[string]interface{}); ok {
				rlt, _ := expr.Eval(govaluate.MapParameters(m))
				switch v := rlt.(type) {
				case float64:
					if math.IsNaN(v) || math.IsInf(v, 0) {
						rlt = nil
					}
				}
				out.Push(rlt)
			} else {
				panic("invalid item")
			}
		}
	},
}
