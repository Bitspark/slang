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
	def, err := readOperatorDef(opFile, nil, nil)

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

// readOperatorDef reads the operator definition for the given file and replaces all generic types according to the
// generics map given. The generics map must not contain any further generic types.
func readOperatorDef(opDefFilePath string, generics map[string]*core.PortDef, pathsRead []string) (core.OperatorDef, error) {
	var def core.OperatorDef

	// Make sure generics is free of further generics
	for _, g := range generics {
		if err := g.FreeOfGenerics(); err != nil {
			return def, err
		}
	}

	// Recursion detection: chick if absolute path is contained in pathsRead
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

	// Read the file
	b, err := ioutil.ReadFile(opDefFilePath)
	if err != nil {
		return def, err
	}

	// Parse the file, just read it in
	def = ParseOperatorDef(string(b))

	// Validate the file
	if !def.Valid() {
		err := def.Validate()
		if err != nil {
			return def, err
		}
	}

	// Replace all generics in the definition
	if err := def.SpecifyGenericPorts(generics); err != nil {
		return def, err
	}

	// Make sure we replaced all generics in the definition
	if err := def.FreeOfGenerics(); err != nil {
		return def, err
	}

	currDir := path.Dir(opDefFilePath)

	// Descend to child operators
	for _, childOpInsDef := range def.Operators {
		childDef, err := getOperatorDef(childOpInsDef, currDir, pathsRead)

		if err != nil {
			return def, err
		}

		if err := childDef.FreeOfGenerics(); err != nil {
			return def, err
		}

		// Save the definition in the instance for the next build step: creating operators and connecting
		childOpInsDef.SetOperatorDef(childDef)
	}

	return def, nil
}

// getOperatorDef tries to get the operator definition from the builtin package or the file system.
func getOperatorDef(insDef *core.InstanceDef, currDir string, pathsRead []string) (core.OperatorDef, error) {
	if builtin.IsRegistered(insDef.Operator) {
		// Case 1: We found it in the builtin package, return
		return builtin.GetOperatorDef(insDef)
	}

	// Case 2: We have to read it from the file system

	var def core.OperatorDef

	relFilePath := strings.Replace(insDef.Operator, ".", "/", -1) + ".json"

	// Check if it is a local operator which has to be found relative to the current operator
	if strings.HasPrefix(insDef.Operator, ".") {
		defFilePath := path.Join(currDir, relFilePath)
		def, err := readOperatorDef(defFilePath, insDef.Generics, pathsRead)

		if err != nil {
			return def, err
		}

		return def, nil
	}

	// These are the paths where we search for operators
	paths := []string{"."}

	// Iterate through the paths and take the first operator we find
	var err error
	for _, p := range paths {
		defFilePath := path.Join(p, relFilePath)

		def, err = readOperatorDef(defFilePath, insDef.Generics, pathsRead)

		if err != nil {
			continue
		}

		// We found an operator, return
		return def, nil
	}

	// We haven't found an operator, return error
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
	} else if builtin.IsRegistered(insDef.Operator) {
		return nil, err
	}
	return buildAndConnectOperator(insDef.Name, insDef.OperatorDef(), par)
}
