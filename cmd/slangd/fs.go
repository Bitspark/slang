package main

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

func EnsureEnvironVar(key string, dfltVal string) string {
	if val := os.Getenv(key); strings.Trim(val, " ") != "" {
		return val
	}
	os.Setenv(key, dfltVal)
	return dfltVal
}

func EnsureDirExists(dir string) (string, error) {
	err := os.MkdirAll(dir, os.ModePerm)
	return dir, err
}

type FileSystemDumperLoader struct {
	paths []string
}

func NewFileSystemDumperLoader() *FileSystemDumperLoader {
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

	return &FileSystemDumperLoader{paths}
}

func (fs *FileSystemDumperLoader) Has(opId uuid.UUID) bool {
	all, err := fs.List()
	return err == nil && funk.Contains(all, opId)
}

func (fs *FileSystemDumperLoader) List() ([]uuid.UUID, error) {
	var outerErr error

	opsFilePathSet := make(map[uuid.UUID]bool)

	for i := len(fs.paths); i > 0; i-- {
		currRootDir := fs.paths[i-1]
		outerErr = filepath.Walk(currRootDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() ||
				strings.HasPrefix(info.Name(), ".") ||
				strings.Contains(info.Name(), "_") ||
				!fs.hasSupportedSuffix(info.Name()) {
				return nil
			}

			opDef, err := fs.readOperatorDef(path, nil)

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

func (fs *FileSystemDumperLoader) Load(opId uuid.UUID) (*core.OperatorDef, error) {
	opDefFilePath, _, err := fs.getFilePathWithFileEnding(strings.Replace(opId.String(), ".", string(filepath.Separator), -1), "")

	if err != nil {
		return nil, err
	}

	return fs.readOperatorDef(opDefFilePath, nil)
}

func (fs *FileSystemDumperLoader) Dump(opDef core.OperatorDef) (uuid.UUID, error) {
	opId, err := uuid.Parse(opDef.Id)

	if err != nil {
		return opId, fmt.Errorf(`id is not a valid UUID v4: "%s" --> "%s"`, opDef.Id, err)
	}

	cwd := fs.workingDir()

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

func (fs *FileSystemDumperLoader) workingDir() string {
	return fs.paths[0]
}

func (fs *FileSystemDumperLoader) hasSupportedSuffix(filePath string) bool {
	return utils.IsJSON(filePath) || utils.IsYAML(filePath)
}

func (fs *FileSystemDumperLoader) getInstanceName(opDefFilePath string) string {
	return strings.TrimSuffix(filepath.Base(opDefFilePath), filepath.Ext(opDefFilePath))
}

func (fs *FileSystemDumperLoader) getFullyQualifiedName(opDefFilePath string) string {
	return strings.TrimSuffix(filepath.Base(opDefFilePath), filepath.Ext(opDefFilePath))
}

func (fs *FileSystemDumperLoader) getFilePathWithFileEnding(relFilePath string, enforcedPath string) (string, string, error) {
	var err error
	relevantPaths := fs.paths
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
func (fs *FileSystemDumperLoader) readOperatorDef(opDefFilePath string, pathsRead []string) (*core.OperatorDef, error) {

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
		childDef, err := fs.getOperatorDef(childOpInsDef, currDir, pathsRead)
		if err != nil {
			return &def, err
		}
		// Save the definition in the instance for the next build step: creating operators and connecting
		childOpInsDef.OperatorDef = *childDef
	}

	return &def, nil
}

func (fs *FileSystemDumperLoader) getOperatorPath(operator string, currDir string) (string, error) {
	relFilePath := strings.Replace(operator, ".", string(filepath.Separator), -1)
	enforcedPath := "" // when != "" --> only search for operatorDef in path *enforcedPath*
	// Check if it is a local operator which has to be found relative to the current operator
	if strings.HasPrefix(operator, ".") {
		enforcedPath = currDir
	}

	var err error
	var opDefFilePath string

	// Iterate through the paths and take the first operator we find
	if opDefFilePath, _, err = fs.getFilePathWithFileEnding(relFilePath, enforcedPath); err == nil {
		return opDefFilePath, nil
	}

	return "", err
}

// getOperatorDef tries to get the operator definition from the builtin package or the file system.
func (fs *FileSystemDumperLoader) getOperatorDef(insDef *core.InstanceDef, currDir string, pathsRead []string) (*core.OperatorDef, error) {
	if elem.IsRegistered(insDef.Operator) {
		// Case 1: We found it in the builtin package, return
		return elem.GetOperatorDef(insDef.Operator)
	}

	// Case 2: We have to read it from the file system
	var def *core.OperatorDef
	var err error
	var opDefFilePath string

	if opDefFilePath, err = fs.getOperatorPath(insDef.Operator, currDir); err == nil {
		if def, err = fs.readOperatorDef(opDefFilePath, pathsRead); err == nil {
			return def, nil
		}
	}

	// We haven't found an operator, return error
	return def, err
}

func (fs *FileSystemDumperLoader) PackOperator(zipWriter *zip.Writer, opId uuid.UUID, read map[uuid.UUID]bool) error {
	// opId already packed
	if r, ok := read[opId]; ok && r {
		return nil
	}

	return fmt.Errorf("not implemeted")

	relPath := strings.Replace(opId.String(), ".", string(filepath.Separator), -1)
	absPath, p, err := fs.getFilePathWithFileEnding(relPath, "")
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
	def, err := fs.readOperatorDef(absPath, nil)
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
			fs.packOperator(zipWriter, ins.Operator, read)
		} else {
			fs.packOperator(zipWriter, baseFqop+ins.Operator[1:], read)
		}
	}
	*/

	return nil
}

/*
// readOperatorDef reads the operator definition for the given file.
func (e *FileSystemDumperLoader) ReadPackageDef(pkgDefFilePath string) (core.PackageDef, error) {
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
