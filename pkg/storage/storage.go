package storage

import (
	"errors"
	"fmt"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/elem"
	"github.com/google/uuid"
)

type Backend interface {
	List() ([]uuid.UUID, error)
	Load(opId uuid.UUID) (*core.OperatorDef, error)
	Has(opId uuid.UUID) bool
}

type WriteableBackend interface {
	Backend
	Save(opDef core.OperatorDef) (uuid.UUID, error)
}

type Storage struct {
	backends []Backend
}

func NewStorage() *Storage {
	return &Storage{make([]Backend, 0)}
}

func (s *Storage) AddBackend(backend Backend) *Storage {
	s.backends = append(s.backends, backend)
	return s
}

func (s *Storage) IsSavedInWritableBackend(opId uuid.UUID) bool {
	writableBackends := s.writeableBackends()
	for _, backend := range writableBackends {
		if backend.Has(opId) {
			return true
		}
	}
	return false
}

func (s *Storage) List() ([]uuid.UUID, error) {
	all := make([]uuid.UUID, 0)

	for _, backend := range s.backends {
		l, err := backend.List()

		if err != nil {
			continue
		}
		all = append(all, l...)
	}

	return all, nil
}

func (s *Storage) Save(opDef core.OperatorDef) (uuid.UUID, error) {
	var opId uuid.UUID
	var err error
	// The question is whether we want multiple backends that are able to take a write
	// because if we need to make sure they all use the same identifier
	writableBackends := s.writeableBackends()
	if len(writableBackends) == 0 {
		return opId, errors.New("No writable backend for saving found")
	}
	for _, backend := range writableBackends {
		opId, err = backend.Save(opDef)
	}
	return opId, err
}

func (s *Storage) writeableBackends() []WriteableBackend {
	writeableBackends := make([]WriteableBackend, 0)

	backends := s.selectBackends(func(b Backend) bool {
		b, ok := b.(WriteableBackend)
		return ok
	})
	for _, backend := range backends {
		writeableBackends = append(writeableBackends, backend.(WriteableBackend))
	}
	return writeableBackends
}

func (s *Storage) Load(opId uuid.UUID) (*core.OperatorDef, error) {
	opDef, err := s.getOpDef(opId)
	if err != nil {
		return nil, err
	}
	cpyOpDef := opDef.Copy(true)
	return &cpyOpDef, nil
}

func (s *Storage) selectBackends(f func(Backend) bool) []Backend {
	selected := make([]Backend, 0)
	for _, backend := range s.backends {
		if f(backend) {
			selected = append(selected, backend)
		}
	}
	return selected
}

func (s *Storage) selectBackend(opId uuid.UUID) Backend {
	backends := s.selectBackends(func(b Backend) bool { return b.Has(opId) })
	if len(backends) > 0 {
		return backends[0] // always return the first backend as there should not be different version of the same operator.
	}
	return nil
}

func (s *Storage) getOpDef(opId uuid.UUID) (*core.OperatorDef, error) {
	if opDef, err := elem.GetOperatorDef(opId); err == nil {
		return opDef, nil
	}

	backend := s.selectBackend(opId)

	if backend == nil {
		return nil, fmt.Errorf("unknown operator for id: %s", opId)
	}

	opDef, err := backend.Load(opId)

	if err != nil {
		return nil, err
	}

	return opDef, nil
}
