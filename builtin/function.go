package builtin

import (
	"errors"
	"slang/op"

	"github.com/Knetic/govaluate"
)

type functionStore struct {
	expr     string
	evalExpr *govaluate.EvaluableExpression
}

func functionCreator(def op.InstanceDef, par *op.Operator) (*op.Operator, error) {
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

	inDef := op.PortDef{
		Type: "map",
		Map:  make(map[string]op.PortDef),
	}

	vars := evalExpr.Vars()

	for _, v := range vars {
		inDef.Map[v] = op.PortDef{Type: "any"}
	}

	outDef := op.PortDef{
		Type: "any",
	}

	o, err := op.MakeOperator(def.Name, func(in, out *op.Port, store interface{}) {
		expr := store.(functionStore).evalExpr
		for true {
			i := in.Pull()
			if m, ok := i.(map[string]interface{}); ok {
				rlt, _ := expr.Evaluate(m)
				out.Push(rlt)
			} else {
				out.Push(i)
			}
		}
	}, inDef, outDef, par)
	o.SetStore(functionStore{expr, evalExpr})

	return o, nil
}
