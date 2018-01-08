package builtin

import (
	"errors"
	"slang/core"
	"strconv"
	"strings"

	"github.com/Knetic/govaluate"
)

type EvaluableExpression struct {
	govaluate.EvaluableExpression
}

func getFlattenedObj(obj interface{}) map[string]interface{} {
	flatMap := make(map[string]interface{})

	if a, ok := obj.([]interface{}); ok {
		for k, val := range a {
			key := strconv.Itoa(k)
			var sub interface{}
			var ok bool

			if sub, ok = val.(map[string]interface{}); !ok {
				if sub, ok = val.([]interface{}); !ok {
					flatMap[key] = val
					continue
				}
			}

			for sKey, sVal := range getFlattenedObj(sub) {
				flatMap[key+"__"+sKey] = sVal
			}
		}

	} else if m, ok := obj.(map[string]interface{}); ok {
		for key, val := range m {
			var sub interface{}
			var ok bool

			if sub, ok = val.(map[string]interface{}); !ok {
				if sub, ok = val.([]interface{}); !ok {
					flatMap[key] = val
					continue
				}
			}

			for sKey, sVal := range getFlattenedObj(sub) {
				flatMap[key+"__"+sKey] = sVal
			}
		}

	} else {
		panic("obj must be list or map")
	}

	return flatMap
}

func NewFlatMapParameters(m map[string]interface{}) govaluate.MapParameters {
	flatMap := getFlattenedObj(m)
	return govaluate.MapParameters(flatMap)
}

func NewEvaluableExpression(expression string) (*EvaluableExpression, error) {
	goEvalExpr, err := govaluate.NewEvaluableExpression(strings.Replace(expression, ".", "__", -1))
	if err == nil {
		return &EvaluableExpression{*goEvalExpr}, nil
	}
	return nil, err
}

type functionStore struct {
	expr     string
	evalExpr *EvaluableExpression
}

var evalOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		In: core.PortDef{
			Type:    "generic",
			Generic: "paramsMap",
		},
		Out: core.PortDef{
			Type: "primitive",
		},
	},
	oFunc: func(in, out *core.Port, store interface{}) {
		expr := store.(functionStore).evalExpr
		for true {
			i := in.Pull()

			if isMarker(i) {
				out.Push(i)
				continue
			}

			if m, ok := i.(map[string]interface{}); ok {
				rlt, _ := expr.Eval(NewFlatMapParameters(m))
				out.Push(rlt)
			} else {
				panic("invalid item")
			}
		}
	},
	oPropFunc: func(o *core.Operator, props map[string]interface{}) error {
		exprStr, ok := props["expression"]

		if !ok {
			return errors.New("no expression given")
		}

		expr, ok := exprStr.(string)

		if !ok {
			return errors.New("expression must be string")
		}

		evalExpr, err := NewEvaluableExpression(expr)

		if err != nil {
			return err
		}

		if o != nil {
			o.SetStore(functionStore{expr, evalExpr})
		}

		return nil
	},
}
