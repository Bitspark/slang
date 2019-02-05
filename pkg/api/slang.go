package api

import (
	"fmt"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/elem"
)

func CreateAndConnectOperator(insName string, def core.OperatorDef, ordered bool) (*core.Operator, error) {
	// Create new non-builtin operator
	o, err := core.NewOperator(insName, nil, nil, nil, nil, def)
	if err != nil {
		return nil, err
	}

	// Recursively create all child operators from top to bottom
	for _, childOpInsDef := range def.InstanceDefs {
		if builtinOp, err := elem.MakeOperator(*childOpInsDef); err == nil {
			// Builtin operator has been found
			builtinOp.SetParent(o)
			continue
		} else if elem.IsRegistered(childOpInsDef.Operator) {
			// Builtin operator with that name exists, but still could not create it, so an error must have occurred
			return nil, err
		}

		oc, err := CreateAndConnectOperator(childOpInsDef.Name, childOpInsDef.OperatorDef, ordered)
		if err != nil {
			return nil, err
		}

		oc.SetParent(o)
	}

	// Parse all connections before starting to connect
	parsedConns := make(map[*core.Port][]*core.Port)
	for srcConnDef, dstConnDefs := range def.Connections {
		if pSrc, err := core.ParsePortReference(srcConnDef, o); err == nil {
			parsedConns[pSrc] = nil
			for _, dstConnDef := range dstConnDefs {
				if pDst, err := core.ParsePortReference(dstConnDef, o); err == nil {
					parsedConns[pSrc] = append(parsedConns[pSrc], pDst)
				} else {
					return nil, fmt.Errorf("%s: %s", err.Error(), dstConnDef)
				}
			}
		} else {
			return nil, fmt.Errorf("%s: %s", err.Error(), srcConnDef)
		}
	}

	if err := connectDestinations(o, parsedConns, ordered); err != nil {
		return nil, err
	}

	return o, nil
}

// connectDestinations connects operators following from the in port to the out port
func connectDestinations(o *core.Operator, conns map[*core.Port][]*core.Port, ordered bool) error {
	var ops []*core.Operator
	for pSrc, pDsts := range conns {
		if pSrc.Operator() != o {
			continue
		}
		// Start with operator o
		for _, pDst := range pDsts {
			if err := pSrc.Connect(pDst); err != nil {
				return fmt.Errorf("%s -> %s: %s", pSrc.Name(), pDst.Name(), err)
			}
			ops = append(ops, pDst.Operator())
		}
		// Set the destinations nil so that we do not end in an infinite recursion
		conns[pSrc] = nil
	}

	var contdOps []*core.Operator
	if ordered {
		// Filter for ops that have all in ports connected
		for _, op := range ops {
			connected := true
			for _, pDsts := range conns {
				for _, pDst := range pDsts {
					if op == pDst.Operator() && pDst.Delegate() == nil {
						connected = false
						goto end
					}
				}
			}
		end:
			if connected {
				contdOps = append(contdOps, op)
			}
		}
	} else {
		contdOps = ops
	}

	// Continue with ops that are completely connected
	for _, op := range contdOps {
		if err := connectDestinations(op, conns, ordered); err != nil {
			return err
		}
	}
	return nil
}

func BuildAndCompile(opDef core.OperatorDef, gens map[string]*core.TypeDef, props map[string]interface{}) (*core.Operator, error) {
	if op, err := Build(opDef, gens, props); err == nil {
		return Compile(op)
	} else {
		return op, err
	}
}

func Build(opDef core.OperatorDef, gens map[string]*core.TypeDef, props map[string]interface{}) (*core.Operator, error) {
	// Recursively replace generics by their actual types and propagate properties
	// TODO SpecifyOperator should instantiate and return an Operator
	opDef = opDef.Copy()
	err := opDef.SpecifyOperator(gens, props)
	if err != nil {
		return nil, err
	}

	// Create and connect the operator
	op, err := CreateAndConnectOperator("", opDef, false)
	if err != nil {
		return nil, err
	}

	return op, nil
}

func Compile(op *core.Operator) (*core.Operator, error) {
	// Compile
	op.Compile()

	// Connect
	flatDef, err := op.Define()
	if err != nil {
		return nil, err
	}

	// Create and connect the flat operator
	flatOp, err := CreateAndConnectOperator("", flatDef, true)
	if err != nil {
		return nil, err
	}

	// Check if all in ports are connected
	err = flatOp.CorrectlyCompiled()
	if err != nil {
		return nil, err
	}

	return flatOp, nil
}
