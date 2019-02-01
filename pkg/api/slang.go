package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/elem"
	"gopkg.in/yaml.v2"
	"strings"
)

func ParseTypeDef(defStr string) core.TypeDef {
	def := core.TypeDef{}
	json.Unmarshal([]byte(defStr), &def)
	return def
}

func ParseJSONOperatorDef(defStr string) (core.OperatorDef, error) {
	def := core.OperatorDef{}
	err := json.Unmarshal([]byte(defStr), &def)
	return def, err
}

func ParseYAMLOperatorDef(defStr string) (core.OperatorDef, error) {
	def := core.OperatorDef{}
	err := yaml.Unmarshal([]byte(defStr), &def)
	for _, id := range def.InstanceDefs {
		id.Properties.Clean()
	}
	return def, err
}

func ParseYAMLPackageDef(defStr string) (core.PackageDef, error) {
	def := core.PackageDef{}
	err := yaml.Unmarshal([]byte(defStr), &def)
	return def, err
}

func ParsePortReference(refStr string, par *core.Operator) (*core.Port, error) {
	if par == nil {
		return nil, errors.New("operator must not be nil")
	}
	if len(refStr) == 0 {
		return nil, errors.New("empty connection string")
	}

	var in bool
	sep := ""
	opIdx := 0
	portIdx := 0
	if strings.Contains(refStr, "(") {
		in = true
		sep = "("
		opIdx = 1
		portIdx = 0
	} else if strings.Contains(refStr, ")") {
		in = false
		sep = ")"
		opIdx = 0
		portIdx = 1
	} else {
		return nil, errors.New("cannot derive direction")
	}

	refSplit := strings.Split(refStr, sep)
	if len(refSplit) != 2 {
		return nil, fmt.Errorf(`connection string malformed (1): "%s"`, refStr)
	}
	opPart := refSplit[opIdx]
	portPart := refSplit[portIdx]

	var o *core.Operator
	var p *core.Port
	if opPart == "" {
		o = par
		if in {
			p = o.Main().In()
		} else {
			p = o.Main().Out()
		}
	} else {
		if strings.Contains(opPart, ".") && strings.Contains(opPart, "@") {
			return nil, fmt.Errorf(`cannot reference both service and delegate: "%s"`, refStr)
		}
		if strings.Contains(opPart, ".") {
			opSplit := strings.Split(opPart, ".")
			if len(opSplit) != 2 {
				return nil, fmt.Errorf(`connection string malformed (2): "%s"`, refStr)
			}
			opName := opSplit[0]
			dlgName := opSplit[1]
			if opName == "" {
				o = par
			} else {
				o = par.Child(opName)
				if o == nil {
					return nil, fmt.Errorf(`operator "%s" has no child "%s"`, par.Name(), opName)
				}
			}
			if dlg := o.Delegate(dlgName); dlg != nil {
				if in {
					p = dlg.In()
				} else {
					p = dlg.Out()
				}
			} else {
				return nil, fmt.Errorf(`operator "%s" has no delegate "%s"`, o.Name(), dlgName)
			}
		} else if strings.Contains(opPart, "@") {
			opSplit := strings.Split(opPart, "@")
			if len(opSplit) != 2 {
				return nil, fmt.Errorf(`connection string malformed (3): "%s"`, refStr)
			}
			opName := opSplit[1]
			srvName := opSplit[0]
			if opName == "" {
				o = par
			} else {
				o = par.Child(opName)
				if o == nil {
					return nil, fmt.Errorf(`operator "%s" has no child "%s"`, par.Name(), opName)
				}
			}
			if srv := o.Service(srvName); srv != nil {
				if in {
					p = srv.In()
				} else {
					p = srv.Out()
				}
			} else {
				return nil, fmt.Errorf(`operator "%s" has no service "%s"`, o.Name(), srvName)
			}
		} else {
			o = par.Child(opPart)
			if o == nil {
				return nil, fmt.Errorf(`operator "%s" has no child "%s"`, par.Name(), refSplit[0])
			}
			if in {
				p = o.Main().In()
			} else {
				p = o.Main().Out()
			}
		}
	}

	pathSplit := strings.Split(portPart, ".")
	if len(pathSplit) == 1 && pathSplit[0] == "" {
		return p, nil
	}

	for i := 0; i < len(pathSplit); i++ {
		if pathSplit[i] == "~" {
			p = p.Stream()
			if p == nil {
				return nil, errors.New("descending too deep (stream)")
			}
			continue
		}

		if p.Type() != core.TYPE_MAP {
			return nil, errors.New("descending too deep (map)")
		}

		k := pathSplit[i]
		p = p.Map(k)
		if p == nil {
			return nil, fmt.Errorf("unknown port: %s", k)
		}
	}

	return p, nil
}

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
		if pSrc, err := ParsePortReference(srcConnDef, o); err == nil {
			parsedConns[pSrc] = nil
			for _, dstConnDef := range dstConnDefs {
				if pDst, err := ParsePortReference(dstConnDef, o); err == nil {
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

func Build(opDef core.OperatorDef, gens map[string]*core.TypeDef, props map[string]interface{}) (*core.Operator, error) {
	// Recursively replace generics by their actual types and propagate properties
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
