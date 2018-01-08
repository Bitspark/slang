package slang

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"slang/builtin"
	"slang/core"
	"strings"
)

func BuildOperator(opFile string, compile bool) (*core.Operator, error) {
	// Read operator definition and perform recursion detection
	def, err := readOperatorDef(opFile, nil)

	if err != nil {
		return nil, err
	}

	// Create and internally connect the operator
	op, err := buildAndConnectOperator(opFile, def, nil)

	if err != nil {
		return nil, err
	}

	if compile {
		// Compile when requested
		op.Compile()
	}

	return op, nil
}

func ParsePortDef(defStr string) core.PortDef {
	def := core.PortDef{}
	json.Unmarshal([]byte(defStr), &def)
	return def
}

func ParseOperatorDef(defStr string) core.OperatorDef {
	def := core.OperatorDef{}
	json.Unmarshal([]byte(defStr), &def)
	return def
}

func ParsePortReference(connStr string, par *core.Operator) (*core.Port, error) {
	if par == nil {
		return nil, errors.New("operator must not be nil")
	}

	if len(connStr) == 0 {
		return nil, errors.New("empty connection string")
	}

	opSplit := strings.Split(connStr, ":")

	if len(opSplit) != 2 {
		return nil, errors.New("connection string malformed")
	}

	var o *core.Operator
	if len(opSplit[0]) == 0 {
		o = par
	} else {
		o = par.Child(opSplit[0])
		if o == nil {
			return nil, fmt.Errorf(`operator "%s" has no child "%s"`, par.Name(), opSplit[0])
		}
	}

	pathSplit := strings.Split(opSplit[1], ".")

	if len(pathSplit) == 0 {
		return nil, errors.New("connection string malformed")
	}

	var p *core.Port
	if pathSplit[0] == "in" {
		p = o.In()
	} else if pathSplit[0] == "out" {
		p = o.Out()
	} else {
		return nil, fmt.Errorf("invalid direction: %s", pathSplit[0])
	}

	for i := 1; i < len(pathSplit); i++ {
		if pathSplit[i] == "" {
			p = p.Stream()
			continue
		}

		if p.Type() != core.TYPE_MAP {
			return nil, errors.New("descending too deep")
		}

		k := pathSplit[i]
		p = p.Map(k)
		if p == nil {
			return nil, fmt.Errorf("unknown port: %s", k)
		}
	}

	return p, nil
}

// READ OPERATOR DEFINITION

func readOperatorDef(opDefFilePath string, pathsRead []string) (core.OperatorDef, error) {
	var def core.OperatorDef

	b, err := ioutil.ReadFile(opDefFilePath)

	if err != nil {
		return def, err
	}

	def = ParseOperatorDef(string(b))

	if !def.Valid() {
		err := def.Validate()
		if err != nil {
			return def, err
		}
	}

	if absPath, err := filepath.Abs(opDefFilePath); err == nil {
		for _, p := range pathsRead {
			if p == absPath {
				return def, fmt.Errorf("recursion in %s", absPath)
			}
		}

		pathsRead = append(pathsRead, absPath)
	} else {
		return def, err
	}

	currDir := path.Dir(opDefFilePath)

	for _, childOpInsDef := range def.Operators {
		childDef, err := getOperatorDef(*childOpInsDef, currDir, pathsRead)

		// Replace any ports in generic operators with according instance port type specifications
		for identifier, pd := range childOpInsDef.Ports {
			if pDef, err := childDef.In.SpecifyAnyPort(identifier, pd); err != nil {
				return def, err
			} else {
				childDef.In = pDef
			}
			if pDef, err  := childDef.Out.SpecifyAnyPort(identifier, pd); err != nil {
				return def, err
			} else {
				childDef.Out = pDef
			}
		}

		if err := childDef.In.FreeOfAnys(); err != nil {
			return def, err
		}
		if err := childDef.Out.FreeOfAnys(); err != nil {
			return def, err
		}

		if err != nil {
			return def, err
		}

		childOpInsDef.SetOperatorDef(childDef)
	}

	return def, nil
}

func getOperatorDef(insDef core.InstanceDef, currDir string, pathsRead []string) (core.OperatorDef, error) {
	if builtin.IsRegistered(insDef.Operator) {
		return builtin.GetOperatorDef(insDef.Operator), nil
	}

	var def core.OperatorDef
	relFilePath := strings.Replace(insDef.Operator, ".", "/", -1) + ".json"

	if strings.HasPrefix(insDef.Operator, ".") {
		defFilePath := path.Join(currDir, relFilePath)
		def, err := readOperatorDef(defFilePath, pathsRead)

		if err != nil {
			return def, err
		}

		return def, nil
	}

	// These are the paths where we search for operators
	paths := []string{"."}

	var err error
	for _, p := range paths {
		defFilePath := path.Join(p, relFilePath)

		def, err = readOperatorDef(defFilePath, pathsRead)

		if err != nil {
			continue
		}

		return def, nil
	}

	return def, err
}

// MAKE OPERATORS, PORTS AND CONNECTIONS

func buildAndConnectOperator(insName string, def core.OperatorDef, par *core.Operator) (*core.Operator, error) {
	if !def.Valid() {
		err := def.Validate()
		if err != nil {
			return nil, err
		}
	}

	o, err := core.NewOperator(insName, nil, def.In, def.Out)
	o.SetParent(par)

	if err != nil {
		return nil, err
	}

	for _, childOpInsDef := range def.Operators {
		_, err := getOperator(*childOpInsDef, o)

		if err != nil {
			return nil, err
		}
	}

	for srcConnDef, dstConnDefs := range def.Connections {
		if pSrc, err := ParsePortReference(srcConnDef, o); err == nil {
			for _, dstConnDef := range dstConnDefs {
				if pDst, err := ParsePortReference(dstConnDef, o); err == nil {
					if err := pSrc.Connect(pDst); err != nil {
						return nil, err
					}
				} else {
					return nil, err
				}
			}
		} else {
			return nil, err
		}
	}

	return o, nil
}

func getOperator(insDef core.InstanceDef, par *core.Operator) (*core.Operator, error) {
	if builtinOp, err := builtin.MakeOperator(insDef); err == nil {
		builtinOp.SetParent(par)
		return builtinOp, nil
	}
	return buildAndConnectOperator(insDef.Name, insDef.OperatorDef(), par)
}
