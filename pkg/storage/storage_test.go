package storage

import (
	"testing"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/google/uuid"
)

func Test_ReadOnlyStorage(t *testing.T) {
	a := assertions.New(t)
	s := NewStorage().AddBackend(NewReadOnlyFileSystem("/somewhere"))
	w := s.writeableBackends()
	a.Empty(w)
}

func Test_WritableStorage(t *testing.T) {
	a := assertions.New(t)
	s := NewStorage()
	s.AddBackend(NewWritableFileSystem("/somewhere"))
	s.AddBackend(NewWritableFileSystem("/somewhere"))
	s.AddBackend(NewReadOnlyFileSystem("/somewhere"))
	s.AddBackend(NewReadOnlyFileSystem("/somewhere"))
	w := s.writeableBackends()
	a.Len(w, 2)
}

func Test_SavingWithoutWritableBackend(t *testing.T) {
	var u uuid.UUID //empty UUID
	a := assertions.New(t)
	s := NewStorage().AddBackend(NewReadOnlyFileSystem("/somewhere"))
	id, err := s.Save(core.Blueprint{})
	a.Equal(id, u)
	a.EqualError(err, "No writable backend for saving found")
}
