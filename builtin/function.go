package builtin

import (
	"slang"
	"errors"
	"github.com/Knetic/govaluate"
)

type functionStore struct {
	expr     string
	evalExpr *govaluate.EvaluableExpression
}

func functionCreator(def slang.InstanceDef) (*slang.Operator, error) {
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

	inDef := slang.PortDef{
		Type: "map",
		Map:  make(map[string]slang.PortDef),
	}

	vars := evalExpr.Vars()

	for _, v := range vars {
		inDef.Map[v] = slang.PortDef{Type: "any"}
	}

	outDef := slang.PortDef{
		Type: "any",
	}

	o, err := slang.MakeOperator("", func(in, out *slang.Port, store interface{}) {
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
	}, inDef, outDef, nil)
	o.SetStore(functionStore{expr, evalExpr})

	return o, nil
}
