package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"strings"
)

func ParseTypeDef(defStr string) TypeDef {
	def := TypeDef{}
	json.Unmarshal([]byte(defStr), &def)
	return def
}

func ParseJSONOperatorDef(defStr string) (OperatorDef, error) {
	def := OperatorDef{}
	err := json.Unmarshal([]byte(defStr), &def)
	return def, err
}

func ParseYAMLOperatorDef(defStr string) (OperatorDef, error) {
	def := OperatorDef{}
	err := yaml.Unmarshal([]byte(defStr), &def)

	for _, id := range def.InstanceDefs {
		id.Properties.Clean()
	}

	for _, tc := range def.TestCases {
		for i, v := range tc.Data.In {
			tc.Data.In[i] = CleanValue(v)
		}
		for i, v := range tc.Data.Out {
			tc.Data.Out[i] = CleanValue(v)
		}
	}

	return def, err
}

func ParsePortReference(refStr string, par *Operator) (*Port, error) {
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

	var o *Operator
	var p *Port
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

		if p.Type() != TYPE_MAP {
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
