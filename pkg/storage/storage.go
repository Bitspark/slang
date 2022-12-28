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
	Load(opId uuid.UUID) (*core.Blueprint, error)
	Has(opId uuid.UUID) bool
}

type WriteableBackend interface {
	Backend
	Save(blueprint core.Blueprint) (uuid.UUID, error)
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

func (s *Storage) Save(blueprint core.Blueprint) (uuid.UUID, error) {
	var opId uuid.UUID
	var err error
	// The question is whether we want multiple backends that are able to take a write
	// because if we need to make sure they all use the same identifier
	writableBackends := s.writeableBackends()
	if len(writableBackends) == 0 {
		return opId, errors.New("no writable backend for saving found")
	}
	for _, backend := range writableBackends {
		opId, err = backend.Save(blueprint)
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

func (s *Storage) Load(opId uuid.UUID) (*core.Blueprint, error) {
	blueprint, err := s.getBlueprintId(opId)
	if err != nil {
		return nil, err
	}
	cpyBlueprint := blueprint.Copy(true)
	return &cpyBlueprint, nil
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

func (s *Storage) getBlueprintId(opId uuid.UUID) (*core.Blueprint, error) {
	if blueprint, err := elem.GetBlueprint(opId); err == nil {
		return blueprint, nil
	}

	backend := s.selectBackend(opId)

	if backend == nil {
		return nil, fmt.Errorf("unknown operator for id: %s", opId)
	}

	blueprint, err := backend.Load(opId)

	if err != nil {
		return nil, err
	}

	return blueprint, nil
}
