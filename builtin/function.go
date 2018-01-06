package builtin

import (
	"errors"
	"slang/core"

	"github.com/Knetic/govaluate"
)

type functionStore struct {
	expr     string
	evalExpr *govaluate.EvaluableExpression
}

func createOpFunc(def core.InstanceDef, par *core.Operator) (*core.Operator, error) {
	if def.Properties == nil {
		return nil, errors.New("no properties given")
	}

	exprStr, ok := def.Properties["expression"]

	if !ok {
		return nil, errors.New("no expression given")
	}

	expr, ok := exprStr.(string)

	if !ok {
		return nil, errors.New("expression must be string")
	}

	evalExpr, err := govaluate.NewEvaluableExpression(expr)

	if err != nil {
		return nil, err
	}

	inDef := core.PortDef{
		Type: "map",
		Map:  make(map[string]core.PortDef),
	}

	vars := evalExpr.Vars()

	for _, v := range vars {
		inDef.Map[v] = core.PortDef{Type: "primitive"}
	}

	outDef := core.PortDef{
		Type: "primitive",
	}

	o, err := core.NewOperator(def.Name, func(in, out *core.Port, store interface{}) {
		expr := store.(functionStore).evalExpr
		for true {
			i := in.Pull()

			if isMarker(i) {
				out.Push(i)
				continue
			}

			if m, ok := i.(map[string]interface{}); ok {
				rlt, _ := expr.Evaluate(m)
				out.Push(rlt)
			} else {
				panic("invalid item")
			}
		}
	}, inDef, outDef, par)
	o.SetStore(functionStore{expr, evalExpr})

	return o, nil
}
