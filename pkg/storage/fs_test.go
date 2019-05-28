package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Bitspark/slang/tests/assertions"
)

func Test_ReadOnlyFilesystem(t *testing.T) {
	a := assertions.New(t)
	fs := NewReadOnlyFileSystem("/somewhere")
	a.Implements((*Backend)(nil), fs)
}
func Test_WriteableFilesystem(t *testing.T) {
	a := assertions.New(t)
	fs := NewWritableFileSystem("/somewhere")
	a.Implements((*Backend)(nil), fs)
	a.Implements((*WriteableBackend)(nil), fs)
}

func Test_cleanPath__AppendSlash(t *testing.T) {
	a := assertions.New(t)
	path := cleanPath("/tmp/folder")
	abspath, err := filepath.Abs("/tmp/folder")
	a.NoError(err)
	a.Equal(abspath+string(filepath.Separator), path)
}

func Test_cleanPath__ExpandRelativePath(t *testing.T) {
	cwd, _ := os.Getwd()
	a := assertions.New(t)
	path := cleanPath("folder")
	// filepath.join stips trailing slash
	a.Equal(filepath.Join(cwd, "folder")+string(filepath.Separator), path)
}
