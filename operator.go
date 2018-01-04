package slang

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"slang/builtin"
	"slang/op"
	"strings"
)

func ReadOperator(opDefFilePath string) (*op.Operator, error) {
	return readOperator(opDefFilePath, opDefFilePath, nil)
}

func readOperator(insName string, opDefFilePath string, par *op.Operator) (*op.Operator, error) {
	b, err := ioutil.ReadFile(opDefFilePath)

	if err != nil {
		return nil, err
	}

	def := getJSONOperatorDef(string(b))

	if !def.Valid() {
		err := def.Validate()
		if err != nil {
			return nil, err
		}
	}

	o, err := op.MakeOperator(insName, nil, *def.In, *def.Out, par)

	if err != nil {
		return nil, err
	}

	currDir := path.Dir(opDefFilePath)

	for _, childOpInsDef := range def.Operators {
		_, err := getOperator(childOpInsDef, o, currDir)

		if err != nil {
			return nil, err
		}
	}

	for srcConnDef, dstConnDefs := range def.Connections {
		if pSrc, err := ParseConnection(srcConnDef, o); err == nil {
			for _, dstConnDef := range dstConnDefs {
				if pDst, err := ParseConnection(dstConnDef, o); err == nil {
					pSrc.Connect(pDst)
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

func getOperator(insDef op.InstanceDef, par *op.Operator, currDir string) (*op.Operator, error) {
	if builtinOp, err := builtin.MakeOperator(insDef, par); err == nil {
		return builtinOp, nil
	}

	relFilePath := strings.Replace(insDef.Operator, ".", "/", -1) + ".json"

	if strings.HasPrefix(insDef.Operator, ".") {
		defFilePath := path.Join(currDir, relFilePath)
		o, err := readOperator(insDef.Name, defFilePath, par)

		if err != nil {
			return nil, err
		}

		return o, nil
	}

	paths := []string{"."}

	var err error
	for _, p := range paths {
		defFilePath := path.Join(p, relFilePath)

		var o *op.Operator
		o, err = readOperator(insDef.Name, defFilePath, par)

		if err != nil {
			continue
		}

		return o, nil
	}

	return nil, err
}

func ParseConnection(connStr string, par *op.Operator) (*op.Port, error) {
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

	var o *op.Operator
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

	var p *op.Port
	if pathSplit[0] == "in" {
		p = o.In()
	} else if pathSplit[0] == "out" {
		p = o.Out()
	} else {
		return nil, fmt.Errorf("invalid direction: %s", pathSplit[1])
	}

	for p.Type() == op.TYPE_STREAM {
		p = p.Stream()
	}

	for i := 1; i < len(pathSplit); i++ {
		if p.Type() != op.TYPE_MAP {
			return nil, errors.New("descending too deep")
		}

		k := pathSplit[i]
		p = p.Map(k)
		if p == nil {
			return nil, fmt.Errorf("unknown port: %s", k)
		}

		for p.Type() == op.TYPE_STREAM {
			p = p.Stream()
		}
	}

	return p, nil
}
