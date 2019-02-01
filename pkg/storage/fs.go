package storage

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/Bitspark/go-funk"
	"github.com/Bitspark/slang/pkg/api"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/elem"
	"github.com/Bitspark/slang/pkg/utils"
	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var FILE_ENDINGS = []string{".yaml", ".yml", ".json"} // Order of endings matters!

func EnsureDirExists(dir string) (string, error) {
	err := os.MkdirAll(dir, os.ModePerm)
	return dir, err
}

type Environ struct {
	paths []string
}

func NewEnviron() *Environ {
	// Stores all library paths in the global paths variable
	// We always look in the local directory first
	paths := []string{}

	missingEnvVars := map[string]bool{}
	// "SLANG_DIR" must be always at beginning
	for _, envName := range []string{"SLANG_DIR", "SLANG_LIB"} {
		envVal := os.Getenv(envName)
		if envVal == "" {
			missingEnvVars[envName] = true
			continue
		}
		paths = append(paths, envVal)
	}

	if len(missingEnvVars) > 0 {
		panic(fmt.Sprintf("environment variable %v is undefined", funk.Keys(missingEnvVars)))
	}

	pathSep := string(filepath.Separator)
	for i, p := range paths {
		p = filepath.Clean(p)
		if !strings.HasSuffix(p, pathSep) {
			p += pathSep
		}
		paths[i] = p
	}

	return &Environ{paths}
}

func NewTestEnviron(cwd string) *Environ {
	os.Setenv("SLANG_LIB", cwd)
	os.Setenv("SLANG_DIR", cwd)
	return NewEnviron()
}

func (e *Environ) IsLibrary(opId uuid.UUID) bool {
	_, _, err := e.getFilePathWithFileEnding(strings.Replace(opId.String(), ".", string(filepath.Separator), -1), e.paths[0])
	return err != nil
}

func (e *Environ) List() ([]uuid.UUID, error) {
	var outerErr error

	opsFilePathSet := make(map[uuid.UUID]bool)

	for i := len(e.paths); i > 0; i-- {
		currRootDir := e.paths[i-1]
		outerErr = filepath.Walk(currRootDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() ||
				strings.HasPrefix(info.Name(), ".") ||
				strings.Contains(info.Name(), "_") ||
				!e.hasSupportedSuffix(info.Name()) {
				return nil
			}

			opDef, err := e.readOperatorDef(path, nil)

			if err != nil {
				return err
			}

			if opId, err := uuid.Parse(opDef.Id); err == nil {
				opsFilePathSet[opId] = true
			} else {
				return err
			}
			return nil
		})
	}

	if outerErr != nil {
		return nil, outerErr
	} else if len(opsFilePathSet) == 0 {
		return nil, fmt.Errorf("no operator files found")
	} else {
		return funk.Keys(opsFilePathSet).([]uuid.UUID), nil
	}
}

func (e *Environ) Load(opId uuid.UUID) (*core.OperatorDef, error) {
	opDefFilePath, _, err := e.getFilePathWithFileEnding(strings.Replace(opId.String(), ".", string(filepath.Separator), -1), "")

	if err != nil {
		return nil, err
	}

	return e.readOperatorDef(opDefFilePath, nil)
}

func (e *Environ) Store(opDef core.OperatorDef) (uuid.UUID, error) {
	opId, err := uuid.Parse(opDef.Id)

	if err != nil {
		return opId, fmt.Errorf(`id is not a valid UUID v4: "%s" --> "%s"`, opDef.Id, err)
	}

	cwd := e.workingDir()

	relPath := strings.Replace(opId.String(), ".", string(filepath.Separator), -1)
	absPath := filepath.Join(cwd, relPath+".yaml")
	_, err = EnsureDirExists(filepath.Dir(absPath))

	if err != nil {
		return opId, err
	}

	opDefYaml, err := yaml.Marshal(&opDef)

	if err != nil {
		return opId, err
	}

	err = ioutil.WriteFile(absPath, opDefYaml, os.ModePerm)
	if err != nil {
		return opId, err
	}

	return opId, nil
}

func (e *Environ) workingDir() string {
	return e.paths[0]
}

/*
func (e *Environ) BuildAndCompileOperator(opFilePath string, gens map[string]*core.TypeDef, props map[string]interface{}) (*core.Operator, error) {
	if !filepath.IsAbs(opFilePath) {
		opFilePath = filepath.Join(e.workingDir(), opFilePath)
	}

	insName := ""

	// Find correct file
	opDefFilePath, err := utils.FileWithFileEnding(opFilePath, FILE_ENDINGS)
	if err != nil {
		return nil, err
	}

	// Recursively read operator definitions and perform recursion detection
	def, err := e.readOperatorDef(opDefFilePath, nil)
	if err != nil {
		return nil, err
	}

	// Recursively replace generics by their actual types and propagate properties
	err = def.SpecifyOperator(gens, props)
	if err != nil {
		return nil, err
	}

	// Create and connect the operator
	op, err := api.CreateAndConnectOperator(insName, def, false)
	if err != nil {
		return nil, err
	}

	// Compile
	op.Compile()

	// Connect
	flatDef, err := op.Define()
	if err != nil {
		return nil, err
	}

	// Create and connect the flat operator
	flatOp, err := api.CreateAndConnectOperator(insName, flatDef, true)
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
*/

func (e *Environ) hasSupportedSuffix(filePath string) bool {
	return utils.IsJSON(filePath) || utils.IsYAML(filePath)
}

func (e *Environ) getInstanceName(opDefFilePath string) string {
	return strings.TrimSuffix(filepath.Base(opDefFilePath), filepath.Ext(opDefFilePath))
}

func (e *Environ) getFullyQualifiedName(opDefFilePath string) string {
	return strings.TrimSuffix(filepath.Base(opDefFilePath), filepath.Ext(opDefFilePath))
}

func (e *Environ) getFilePathWithFileEnding(relFilePath string, enforcedPath string) (string, string, error) {
	var err error
	relevantPaths := e.paths
	if enforcedPath != "" {
		relevantPaths = []string{enforcedPath}
	}

	for _, p := range relevantPaths {
		defFilePath := filepath.Join(p, relFilePath)
		// Find correct file
		var opDefFilePath string

		opDefFilePath, err = utils.FileWithFileEnding(defFilePath, FILE_ENDINGS)
		if err != nil {
			continue
		}

		return opDefFilePath, p, nil
	}

	return "", "", err
}

// readOperatorDef reads the operator definition for the given file.
func (e *Environ) readOperatorDef(opDefFilePath string, pathsRead []string) (*core.OperatorDef, error) {

	b, err := ioutil.ReadFile(opDefFilePath)
	if err != nil {
		return nil, errors.New("could not read operator file " + opDefFilePath)
	}

	// Recursion detection: chick if absolute path is contained in pathsRead
	if absPath, err := filepath.Abs(opDefFilePath); err == nil {
		for _, p := range pathsRead {
			if p == absPath {
				return nil, fmt.Errorf("recursion in %s", absPath)
			}
		}

		pathsRead = append(pathsRead, absPath)
	} else {
		return nil, err
	}

	var def core.OperatorDef
	// Parse the file, just read it in
	if utils.IsYAML(opDefFilePath) {
		def, err = api.ParseYAMLOperatorDef(string(b))
	} else if utils.IsJSON(opDefFilePath) {
		def, err = api.ParseJSONOperatorDef(string(b))
	} else {
		err = errors.New("unsupported file ending")
	}
	if err != nil {
		return nil, err
	}

	// Validate the file
	if !def.Valid() {
		err := def.Validate()
		if err != nil {
			return &def, err
		}
	}

	currDir := filepath.Dir(opDefFilePath)

	// Descend to child operators
	for _, childOpInsDef := range def.InstanceDefs {
		childDef, err := e.getOperatorDef(childOpInsDef, currDir, pathsRead)
		if err != nil {
			return &def, err
		}
		// Save the definition in the instance for the next build step: creating operators and connecting
		childOpInsDef.OperatorDef = *childDef
	}

	return &def, nil
}

func (e *Environ) getOperatorPath(operator string, currDir string) (string, error) {
	relFilePath := strings.Replace(operator, ".", string(filepath.Separator), -1)
	enforcedPath := "" // when != "" --> only search for operatorDef in path *enforcedPath*
	// Check if it is a local operator which has to be found relative to the current operator
	if strings.HasPrefix(operator, ".") {
		enforcedPath = currDir
	}

	var err error
	var opDefFilePath string

	// Iterate through the paths and take the first operator we find
	if opDefFilePath, _, err = e.getFilePathWithFileEnding(relFilePath, enforcedPath); err == nil {
		return opDefFilePath, nil
	}

	return "", err
}

// getOperatorDef tries to get the operator definition from the builtin package or the file system.
func (e *Environ) getOperatorDef(insDef *core.InstanceDef, currDir string, pathsRead []string) (*core.OperatorDef, error) {
	if elem.IsRegistered(insDef.Operator) {
		// Case 1: We found it in the builtin package, return
		return elem.GetOperatorDef(insDef.Operator)
	}

	// Case 2: We have to read it from the file system
	var def *core.OperatorDef
	var err error
	var opDefFilePath string

	if opDefFilePath, err = e.getOperatorPath(insDef.Operator, currDir); err == nil {
		if def, err = e.readOperatorDef(opDefFilePath, pathsRead); err == nil {
			return def, nil
		}
	}

	// We haven't found an operator, return error
	return def, err
}

func (e *Environ) PackOperator(zipWriter *zip.Writer, opId uuid.UUID, read map[uuid.UUID]bool) error {
	// opId already packed
	if r, ok := read[opId]; ok && r {
		return nil
	}

	return fmt.Errorf("not implemeted")

	relPath := strings.Replace(opId.String(), ".", string(filepath.Separator), -1)
	absPath, p, err := e.getFilePathWithFileEnding(relPath, "")
	if err != nil {
		return err
	}

	fileWriter, _ := zipWriter.Create(filepath.ToSlash(absPath[len(p):]))
	fileContents, err := ioutil.ReadFile(absPath)
	if err != nil {
		return err
	}
	fileWriter.Write(fileContents)

	read[opId] = true

	/* TODO var suffixes = []string{"_visual.yaml"}
	var absBasePath string
	if strings.HasSuffix(absPath, ".yaml") {
		absBasePath = absPath[:len(absPath)-len(".yaml")]
	} else if strings.HasSuffix(absPath, ".json") {
		absBasePath = absPath[:len(absPath)-len(".json")]
	}

	for _, suffix := range suffixes {
		fileContents, err := ioutil.ReadFile(absBasePath + suffix)
		if err != nil {
			continue
		}
		fileWriter, _ := zipWriter.Create(filepath.ToSlash(absBasePath[len(p):] + suffix))
		fileWriter.Write(fileContents)
	}
	* /

	/*
	def, err := e.readOperatorDef(absPath, nil)
	if err != nil {
		return err
	}

	var baseFqop string
	dotIdx := strings.LastIndex(opId.String(), ".")
	if dotIdx != -1 {
		baseFqop = opId.String()[:dotIdx+1]
	}
	for _, ins := range def.InstanceDefs {
		if strings.HasPrefix(ins.Operator, "slang.") {
			continue
		}
		if elem.IsRegistered(ins.Operator) {
			continue
		}
		if !strings.HasPrefix(ins.Operator, ".") {
			e.packOperator(zipWriter, ins.Operator, read)
		} else {
			e.packOperator(zipWriter, baseFqop+ins.Operator[1:], read)
		}
	}
	*/

	return nil
}

/*
// readOperatorDef reads the operator definition for the given file.
func (e *Environ) ReadPackageDef(pkgDefFilePath string) (core.PackageDef, error) {
	var def core.PackageDef

	b, err := ioutil.ReadFile(pkgDefFilePath)
	if err != nil {
		return def, errors.New("could not read operator file " + pkgDefFilePath)
	}

	// Parse the file, just read it in
	if utils.IsYAML(pkgDefFilePath) {
		def, err = ParseYAMLPackageDef(string(b))
	} else {
		err = errors.New("unsupported file ending")
	}
	if err != nil {
		return def, err
	}

	return def, nil
}
*/
