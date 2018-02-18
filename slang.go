package slang

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
	"path/filepath"
	"slang/builtin"
	"slang/core"
	"slang/utils"
	"strings"
)

var fileEndings = []string{".yaml", ".json"} // Order of endings matters!

func BuildOperator(opFilePath string, props map[string]interface{}, compile bool) (*core.Operator, error) {
	// Find correct file
	opDefFilePath, err := utils.FileWithFileEnding(opFilePath, fileEndings)
	if err != nil {
		return nil, err
	}

	// Read operator definition and perform recursion detection
	def, err := readOperatorDef(opDefFilePath, nil, nil)

	if err != nil {
		return nil, err
	}

	// Create and internally connect the operator
	op, err := buildAndConnectOperator(opFilePath, props, def, nil)

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

func ParseJSONOperatorDef(defStr string) (core.OperatorDef, error) {
	def := core.OperatorDef{}
	err := json.Unmarshal([]byte(defStr), &def)
	return def, err
}

func ParseYAMLOperatorDef(defStr string) (core.OperatorDef, error) {
	def := core.OperatorDef{}
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

	opSplit := strings.Split(refStr, sep)

	if len(opSplit) != 2 {
		return nil, fmt.Errorf(`connection string malformed: ""%s"`, refStr)
	}

	var o *core.Operator
	if opSplit[opIdx] == "" {
		o = par
	} else {
		o = par.Child(opSplit[opIdx])
		if o == nil {
			return nil, fmt.Errorf(`operator "%s" has no child "%s"`, par.Name(), opSplit[0])
		}
	}

	pathSplit := strings.Split(opSplit[portIdx], ".")

	var p *core.Port
	if in {
		p = o.In()
	} else {
		p = o.Out()
	}

	start := 0
	if pathSplit[0] == "" {
		start = 1
	}

	for i := start; i < len(pathSplit); i++ {
		if pathSplit[i] == "" {
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

// READ OPERATOR DEFINITION

// readOperatorDef reads the operator definition for the given file and replaces all generic types according to the
// generics map given. The generics map must not contain any further generic types.
func readOperatorDef(opDefFilePath string, generics map[string]*core.PortDef, pathsRead []string) (core.OperatorDef, error) {
	var def core.OperatorDef

	// Make sure generics is free of further generics
	for _, g := range generics {
		if err := g.GenericsSpecified(); err != nil {
			return def, err
		}
	}

	b, err := ioutil.ReadFile(opDefFilePath)
	if err != nil {
		return def, errors.New("could not read operator file " + opDefFilePath)
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

	// Parse the file, just read it in
	if strings.HasSuffix(opDefFilePath, ".yaml") || strings.HasSuffix(opDefFilePath, ".yml") {
		def, err = ParseYAMLOperatorDef(string(b))
	} else if strings.HasSuffix(opDefFilePath, ".json") {
		def, err = ParseJSONOperatorDef(string(b))
	} else {
		err = errors.New("unsupported file ending")
	}
	if err != nil {
		return def, err
	}

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
	if err := def.GenericsSpecified(); err != nil {
		return def, err
	}

	currDir := path.Dir(opDefFilePath)

	// Descend to child operators
	for _, childOpInsDef := range def.Operators {
		childDef, err := getOperatorDef(childOpInsDef, currDir, pathsRead)

		if err != nil {
			return def, err
		}

		if err := childDef.GenericsSpecified(); err != nil {
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

	relFilePath := strings.Replace(insDef.Operator, ".", "/", -1)

	// Check if it is a local operator which has to be found relative to the current operator
	if strings.HasPrefix(insDef.Operator, ".") {
		defFilePath := path.Join(currDir, relFilePath)
		// Find correct file
		opDefFilePath, err := utils.FileWithFileEnding(defFilePath, fileEndings)
		if err != nil {
			return def, err
		}
		def, err := readOperatorDef(opDefFilePath, insDef.Generics, pathsRead)

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
		// Find correct file
		opDefFilePath, err := utils.FileWithFileEnding(defFilePath, fileEndings)
		if err != nil {
			return def, err
		}
		def, err = readOperatorDef(opDefFilePath, insDef.Generics, pathsRead)

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

// buildAndConnectOperator creates a new non-builtin operator and attaches it to the given parent operator.
// It recursively creates child operators which might as well be builtin operators. It also connects the operators
// according to the given operator definition.
func buildAndConnectOperator(insName string, props map[string]interface{}, def core.OperatorDef, par *core.Operator) (*core.Operator, error) {
	if !def.Valid() {
		err := def.Validate()
		if err != nil {
			return nil, err
		}
	}

	// Check if all properties are defined
	for _, prop := range def.Properties {
		if _, ok := props[prop]; !ok {
			return nil, fmt.Errorf("property \"%s\" has not been specified", prop)
		}
	}

	// Create new non-builtin operator
	o, err := core.NewOperator(insName, nil, def.In, def.Out)
	if err != nil {
		return nil, err
	}

	// Attach it to the parent
	o.SetParent(par)

	// Recursively create all child operators from top to bottom
	for _, childOpInsDef := range def.Operators {
		// Propagate property values to child operators
		for prop, propVal := range childOpInsDef.Properties {
			propKey, ok := propVal.(string)
			if !ok {
				continue
			}
			// Parameterized properties must start with a '$'
			if !strings.HasPrefix(propKey, "$") {
				continue
			}
			propKey = propKey[1:]
			if val, ok := props[propKey]; ok {
				childOpInsDef.Properties[prop] = val
			} else {
				return nil, fmt.Errorf("unknown property \"%s\"", prop)
			}
		}

		_, err := getOperator(*childOpInsDef, o)

		if err != nil {
			return nil, err
		}
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

	if err := connectDestinations(o, parsedConns); err != nil {
		return nil, err
	}

	return o, nil
}

// connectDestinations connects operators following from the in port to the out port
func connectDestinations(o *core.Operator, conns map[*core.Port][]*core.Port) error {
	for pSrc, pDsts := range conns {
		if pSrc.Operator() == o {
			var ops []*core.Operator
			for _, pDst := range pDsts {
				if err := pSrc.Connect(pDst); err != nil {
					return err
				}
				ops = append(ops, pDst.Operator())
			}
			// Set the destinations nil so that we do not end in an infinite recursion
			conns[pSrc] = nil
			for _, op := range ops {
				if err := connectDestinations(op, conns); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// getOperator creates and returns an operator according to the instance definition. If there exists a suitable
// builtin operator it will be returned. Otherwise, a new operator will be created.
func getOperator(insDef core.InstanceDef, par *core.Operator) (*core.Operator, error) {
	if builtinOp, err := builtin.MakeOperator(insDef); err == nil {
		// Builtin operator has been found
		builtinOp.SetParent(par)
		return builtinOp, nil
	} else if builtin.IsRegistered(insDef.Operator) {
		return nil, err
	}
	// Instance definition must have an appropriate operator definition
	if !insDef.OperatorDef().Valid() {
		return nil, errors.New("instance has no operator definition")
	}
	// No builtin operator, so create new one according to the operator definition saved in the instance definition
	return buildAndConnectOperator(insDef.Name, insDef.Properties, insDef.OperatorDef(), par)
}
