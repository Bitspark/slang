package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
	"path/filepath"
	"github.com/Bitspark/slang/pkg/builtin"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/utils"
	"strings"
	"os"
	"runtime"
)

var FILE_ENDINGS = []string{".yaml", ".json"} // Order of endings matters!

type Environ struct {
	paths []string
}

func NewEnviron(workingDir string) *Environ {
	// Stores all library paths in the global paths variable
	// We always look in the local directory first
	paths := []string{workingDir}

	sep := ":"
	if runtime.GOOS == "windows" {
		sep = ";"
	}

	// Read from environment
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		if pair[0] == "SLANG_LIB" {
			libPaths := strings.Split(pair[1], sep)
			for _, libPath := range libPaths {
				if !strings.HasSuffix(libPath, "/") {
					libPath += "/"
				}
				paths = append(paths, libPath)
			}
		}
	}

	return &Environ{paths}
}

func (e *Environ) WorkingDir() string {
	return e.paths[0]
}

func (e *Environ) BuildOperator(opFilePath string, gens map[string]*core.TypeDef, props map[string]interface{}, compile bool) (*core.Operator, error) {
	if !path.IsAbs(opFilePath) {
		opFilePath = path.Join(e.WorkingDir(), opFilePath)
	}

	// Find correct file
	opDefFilePath, err := utils.FileWithFileEnding(opFilePath, FILE_ENDINGS)
	if err != nil {
		return nil, err
	}

	// Read operator definition and perform recursion detection
	def, err := e.ReadOperatorDef(opDefFilePath, nil)

	if err != nil {
		return nil, err
	}

	// Create and internally connect the operator
	op, err := BuildAndConnectOperator(opFilePath, gens, props, def, nil)

	if err != nil {
		return nil, err
	}

	if compile {
		// Compile when requested
		op.Compile()
	}

	return op, nil
}

func ParsePortDef(defStr string) core.TypeDef {
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
			if opSplit[0] == "" {
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
				o = par.Child(opSplit[1])
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

func (e *Environ) getOperatorDefFilePath(relFilePath string, enforcedPath string) (string, error) {

	var err error

	relevantPaths := e.paths
	if enforcedPath != "" {
		relevantPaths = []string{enforcedPath}
	}

	for _, p := range relevantPaths {
		defFilePath := path.Join(p, relFilePath)
		// Find correct file
		var opDefFilePath string

		opDefFilePath, err = utils.FileWithFileEnding(defFilePath, FILE_ENDINGS)

		if err != nil {
			continue
		}

		return opDefFilePath, nil
	}

	return "", err
}

// READ OPERATOR DEFINITION

// ReadOperatorDef reads the operator definition for the given file.
func (e *Environ) ReadOperatorDef(opDefFilePath string, pathsRead []string) (core.OperatorDef, error) {
	var def core.OperatorDef

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

	currDir := path.Dir(opDefFilePath)

	// Descend to child operators
	for _, childOpInsDef := range def.InstanceDefs {
		childDef, err := e.getOperatorDef(childOpInsDef, currDir, pathsRead)

		if err != nil {
			return def, err
		}

		// Save the definition in the instance for the next build step: creating operators and connecting
		childOpInsDef.SetOperatorDef(childDef)
	}

	return def, nil
}

// getOperatorDef tries to get the operator definition from the builtin package or the file system.
func (e *Environ) getOperatorDef(insDef *core.InstanceDef, currDir string, pathsRead []string) (core.OperatorDef, error) {
	if builtin.IsRegistered(insDef.Operator) {
		// Case 1: We found it in the builtin package, return
		return builtin.GetOperatorDef(insDef)
	}

	// Case 2: We have to read it from the file system

	var def core.OperatorDef
	var err error
	var opDefFilePath string

	relFilePath := strings.Replace(insDef.Operator, ".", "/", -1)
	enforcedPath := "" // when != "" --> only search for operatordef in path *enforcedPath*
	// Check if it is a local operator which has to be found relative to the current operator
	if strings.HasPrefix(insDef.Operator, ".") {
		enforcedPath = currDir
	}

	// Iterate through the paths and take the first operator we find
	if opDefFilePath, err = e.getOperatorDefFilePath(relFilePath, enforcedPath); err == nil {
		if def, err = e.ReadOperatorDef(opDefFilePath, pathsRead); err == nil {
			return def, nil
		}
	}

	// We haven't found an operator, return error
	return def, err
}

// MAKE OPERATORS, PORTS AND CONNECTIONS

// BuildAndConnectOperator creates a new non-builtin operator and attaches it to the given parent operator.
// It recursively creates child operators which might as well be builtin operators. It also connects the operators
// according to the given operator definition.
func BuildAndConnectOperator(insName string, gens core.Generics, props core.Properties, def core.OperatorDef, par *core.Operator) (*core.Operator, error) {
	if !def.Valid() {
		err := def.Validate()
		if err != nil {
			return nil, err
		}
	}

	// Create new non-builtin operator
	o, err := core.NewOperator(insName, nil, nil, gens, props, def)
	if err != nil {
		return nil, err
	}

	// Attach it to the parent
	o.SetParent(par)

	// Recursively create all child operators from top to bottom
	for _, childOpInsDef := range def.InstanceDefs {
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

		// TODO: Look at this again
		for _, gen := range childOpInsDef.Generics {
			gen.SpecifyGenerics(gens)
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

	if err := connectAllDestinations(o, parsedConns); err != nil {
		return nil, err
	}

	return o, nil
}

func connectAllDestinations(o *core.Operator, conns map[*core.Port][]*core.Port) error {
	if err := connectDestinations(o, conns, false); err != nil {
		return err
	}
	return connectDestinations(o, conns, true)
}

// connectDestinations connects operators following from the in port to the out port
func connectDestinations(o *core.Operator, conns map[*core.Port][]*core.Port, followDelegates bool) error {
	var ops []*core.Operator
	for pSrc, pDsts := range conns {
		if pSrc.Operator() == o && (followDelegates || pSrc.Delegate() == nil) {
			for _, pDst := range pDsts {
				if err := pSrc.Connect(pDst); err != nil {
					return err
				}
				ops = append(ops, pDst.Operator())
			}
			// Set the destinations nil so that we do not end in an infinite recursion
			conns[pSrc] = nil
		}
	}
	for _, op := range ops {
		if err := connectAllDestinations(op, conns); err != nil {
			return err
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
		// Builtin operator with that name exists, but still could not create it, so an error must have occurred
		return nil, err
	}
	// Instance definition must have an appropriate operator definition
	if !insDef.OperatorDef().Valid() {
		return nil, errors.New("instance has no operator definition")
	}
	// No builtin operator, so create new one according to the operator definition saved in the instance definition
	return BuildAndConnectOperator(insDef.Name, insDef.Generics, insDef.Properties, insDef.OperatorDef(), par)
}
