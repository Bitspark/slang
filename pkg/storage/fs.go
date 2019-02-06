package storage

import (
	"errors"
	"fmt"
	"github.com/Bitspark/go-funk"
	"github.com/Bitspark/slang/pkg/core"
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

type FileSystem struct {
	root string
}

func NewFileSystem(p string) *FileSystem {
	pathSep := string(filepath.Separator)
	p = filepath.Clean(p)
	p, _ = filepath.Abs(p)
	if !strings.HasSuffix(p, pathSep) {
		p += pathSep
	}
	return &FileSystem{p}
}

func (fs *FileSystem) Has(opId uuid.UUID) bool {
	all, err := fs.List()
	return err == nil && funk.Contains(all, opId)
}

func (fs *FileSystem) List() ([]uuid.UUID, error) {
	var outerErr error

	opsFilePathSet := make(map[uuid.UUID]bool)

	outerErr = filepath.Walk(fs.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Prevent recursive walk. Just read files within fs.root
		if info.IsDir() && path != fs.root {
			return filepath.SkipDir
		}

		if info.IsDir() ||
			strings.HasPrefix(info.Name(), ".") ||
			!fs.hasSupportedSuffix(info.Name()) {
			return nil
		}

		opDef, err := fs.readOpDefFile(path)

		if err != nil {
			fmt.Println("-->    ", path, "|", info.Name(), "|", err)
			return err
		}

		if opId, err := uuid.Parse(opDef.Id); err == nil {
			opsFilePathSet[opId] = true
		} else {
			return err
		}
		return nil
	})

	if outerErr != nil {
		return nil, outerErr
	} else if len(opsFilePathSet) == 0 {
		return nil, fmt.Errorf("no operator files found")
	} else {
		return funk.Keys(opsFilePathSet).([]uuid.UUID), nil
	}
}

func (fs *FileSystem) Load(opId uuid.UUID) (*core.OperatorDef, error) {
	opDefFile, err := fs.getFilePath(opId)
	if err != nil {
		return nil, err
	}
	return fs.readOpDefFile(opDefFile)
}

func (fs *FileSystem) Dump(opDef core.OperatorDef) (uuid.UUID, error) {
	opId, err := uuid.Parse(opDef.Id)

	if err != nil {
		return opId, fmt.Errorf(`id is not a valid UUID v4: "%s" --> "%s"`, opDef.Id, err)
	}

	cwd := fs.root

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

func (fs *FileSystem) hasSupportedSuffix(filePath string) bool {
	return utils.IsJSON(filePath) || utils.IsYAML(filePath)
}

func (fs *FileSystem) getInstanceName(opDefFilePath string) string {
	return strings.TrimSuffix(filepath.Base(opDefFilePath), filepath.Ext(opDefFilePath))
}

func (fs *FileSystem) getFilePath(opId uuid.UUID) (string, error) {
	return utils.FileWithFileEnding(filepath.Join(fs.root, opId.String()), FILE_ENDINGS)
}

func (fs *FileSystem) readOpDefFile(opDefFile string) (*core.OperatorDef, error) {
	b, err := ioutil.ReadFile(opDefFile)
	if err != nil {
		return nil, errors.New("could not read operator file " + opDefFile)
	}

	var def core.OperatorDef
	// Parse the file, just read it in
	if utils.IsYAML(opDefFile) {
		def, err = core.ParseYAMLOperatorDef(string(b))
	} else if utils.IsJSON(opDefFile) {
		def, err = core.ParseJSONOperatorDef(string(b))
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
	return &def, nil
}
