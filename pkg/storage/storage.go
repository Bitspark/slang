package storage

import (
	"fmt"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/elem"
	"github.com/google/uuid"
)

type Loader interface {
	List() ([]uuid.UUID, error)
	Has(opId uuid.UUID) bool
	Load(opId uuid.UUID) (*core.OperatorDef, error)
}

type LoaderDumper interface {
	List() ([]uuid.UUID, error)
	Load(opId uuid.UUID) (*core.OperatorDef, error)
	Dump(opDef core.OperatorDef) (uuid.UUID, error)
	Has(opId uuid.UUID) bool
}

type Storage struct {
	loader []Loader
	dumper LoaderDumper
}

func NewStorage(ld LoaderDumper) *Storage {
	st := &Storage{make([]Loader, 0), nil}

	if ld != nil {
		st.dumper = ld
		st.AddLoader(ld)
	}

	return st
}

func (s *Storage) AddLoader(loader Loader) *Storage {
	s.loader = append(s.loader, loader)
	return s
}

func (s *Storage) IsDumpable(opId uuid.UUID) bool {
	return s.dumper.Has(opId)
}

func (s *Storage) IsLoadable(opId uuid.UUID) bool {
	loader := s.findRelatedLoader(opId)
	return loader != nil
}

func (s *Storage) List() ([]uuid.UUID, error) {
	all := make([]uuid.UUID, 0)

	for _, loader := range s.loader {
		l, err := loader.List()

		if err != nil {
			continue
		}
		all = append(all, l...)
	}

	return all, nil
}

func (s *Storage) Store(opDef core.OperatorDef) (uuid.UUID, error) {
	return s.dumper.Dump(opDef)
}

func (s *Storage) Load(opId uuid.UUID) (*core.OperatorDef, error) {
	opDef, err := s.loadFirstFound(opId)
	if err != nil {
		return nil, err
	}
	cpyOpDef := opDef.Copy(true)
	return &cpyOpDef, nil
}

func (s *Storage) findRelatedLoader(opId uuid.UUID) Loader {
	for _, loader := range s.loader {
		if loader.Has(opId) {
			return loader
		}
	}
	return nil
}

func (s *Storage) loadFirstFound(opId uuid.UUID) (*core.OperatorDef, error) {
	if opDef, err := elem.GetOperatorDef(opId.String()); err == nil {
		return opDef, nil
	}

	loader := s.findRelatedLoader(opId)

	if loader == nil {
		return nil, fmt.Errorf("unknown operator for id: %s", opId)
	}

	opDef, err := loader.Load(opId)

	if err != nil {
		return nil, err
	}

	return opDef, nil
}
