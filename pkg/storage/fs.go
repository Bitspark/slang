package storage

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bitspark/go-funk"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/utils"
	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
)

var FILE_ENDINGS = []string{".yaml", ".yml", ".json"} // Order of endings matters!

type FileSystem struct {
	root  string
	cache map[uuid.UUID]*core.Blueprint
	uuids []uuid.UUID
}

type WritableFileSystem struct {
	FileSystem
}

func cleanPath(p string) string {
	pathSep := string(filepath.Separator)
	p = filepath.Clean(p)
	p, _ = filepath.Abs(p)
	if !strings.HasSuffix(p, pathSep) {
		p += pathSep
	}
	return p
}

func NewWritableFileSystem(root string) *WritableFileSystem {
	p := cleanPath(root)
	return &WritableFileSystem{FileSystem: FileSystem{p, nil, nil}}
}

func NewReadOnlyFileSystem(root string) *FileSystem {
	p := cleanPath(root)
	return &FileSystem{p, make(map[uuid.UUID]*core.Blueprint), nil}
}

func (fs *FileSystem) Has(opId uuid.UUID) bool {
	all, err := fs.List()
	return err == nil && funk.Contains(all, opId)
}

func (fs *FileSystem) List() ([]uuid.UUID, error) {
	if fs.uuids != nil {
		return fs.uuids, nil
	}

	opsFilePathSet := make(map[uuid.UUID]bool)

	_ = filepath.Walk(fs.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("cannot read file %s: %s", path, err)
			return nil
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

		blueprint, err := fs.readBlueprintFile(path)

		if err != nil {
			log.Printf("cannot read file %s: %s", path, err)
			return nil
		}

		opsFilePathSet[blueprint.Id] = true

		return nil
	})

	fs.uuids = funk.Keys(opsFilePathSet).([]uuid.UUID)

	return fs.List()
}

func (fs *FileSystem) Load(opId uuid.UUID) (*core.Blueprint, error) {
	if def, ok := fs.cache[opId]; ok {
		return def, nil
	}

	blueprintFile, err := fs.getFilePath(opId)
	if err != nil {
		return nil, err
	}

	fs.cache[opId], err = fs.readBlueprintFile(blueprintFile)
	if err != nil {
		return nil, err
	}

	return fs.Load(opId)
}

func (fs *WritableFileSystem) Save(blueprint core.Blueprint) (uuid.UUID, error) {
	opId := blueprint.Id
	cwd := fs.root
	relPath := strings.Replace(opId.String(), ".", string(filepath.Separator), -1)
	absPath := filepath.Join(cwd, relPath+".yaml")
	_, err := utils.EnsureDirExists(filepath.Dir(absPath))

	if err != nil {
		return opId, err
	}

	delete(fs.cache, opId)
	fs.uuids = append(fs.uuids, opId)

	blueprintYaml, err := yaml.Marshal(&blueprint)

	if err != nil {
		return opId, err
	}

	err = ioutil.WriteFile(absPath, blueprintYaml, os.ModePerm)
	if err != nil {
		return opId, err
	}

	return opId, nil
}

func (fs *WritableFileSystem) List() ([]uuid.UUID, error) {
	// force to reload writable/local blueprints
	fs.clearCache()
	return fs.FileSystem.List()
}

func (fs *WritableFileSystem) Load(opId uuid.UUID) (*core.Blueprint, error) {
	// force to reload writable/local blueprints
	delete(fs.cache, opId)
	return fs.FileSystem.Load(opId)
}

func (fs *WritableFileSystem) clearCache() {
	fs.cache = make(map[uuid.UUID]*core.Blueprint)
	fs.uuids = nil
}

func (fs *FileSystem) hasSupportedSuffix(filePath string) bool {
	return utils.IsJSON(filePath) || utils.IsYAML(filePath)
}

func (fs *FileSystem) getInstanceName(blueprintFilePath string) string {
	return strings.TrimSuffix(filepath.Base(blueprintFilePath), filepath.Ext(blueprintFilePath))
}

func (fs *FileSystem) getFilePath(opId uuid.UUID) (string, error) {
	return utils.FileWithFileEnding(filepath.Join(fs.root, opId.String()), FILE_ENDINGS)
}

func (fs *FileSystem) readBlueprintFile(blueprintFile string) (*core.Blueprint, error) {
	b, err := ioutil.ReadFile(blueprintFile)
	if err != nil {
		return nil, errors.New("could not read operator file " + blueprintFile)
	}

	var def core.Blueprint
	// Parse the file, just read it in
	if utils.IsYAML(blueprintFile) {
		def, err = core.ParseYAMLOperatorDef(string(b))
	} else if utils.IsJSON(blueprintFile) {
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
