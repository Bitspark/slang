package api

import (
	"fmt"
	"github.com/Bitspark/go-funk"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/elem"
	"github.com/google/uuid"
)

type Loader interface {
	List() ([]uuid.UUID, error)
	Has(opId uuid.UUID) bool
	Load(opId uuid.UUID) (*core.OperatorDef, error)
}

type Dumper interface {
	Has(opId uuid.UUID) bool
	Dump(opDef core.OperatorDef) (uuid.UUID, error)
}

type Storage struct {
	loader []Loader
	dumper Dumper
}

func NewStorage(dumper Dumper) *Storage {
	return &Storage{make([]Loader, 0), dumper}
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
			return all, err
		}
		all = append(all, l...)
	}

	return all, nil
}

func (s *Storage) Store(opDef core.OperatorDef) (uuid.UUID, error) {
	return s.dumper.Dump(opDef)
}

func (s *Storage) Load(opId uuid.UUID) (*core.OperatorDef, error) {
	opDef, err := s.load(opId, []string{})
	if err != nil {
		return nil, err
	}
	cpyOpDef := opDef.Copy()
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

func (s *Storage) load(opId uuid.UUID, dependenyChain []string) (*core.OperatorDef, error) {

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

	dependenyChain = append(dependenyChain, opId.String())

	for _, childInsDef := range opDef.InstanceDefs {
		if childInsDef.OperatorDef.Id != "" {
			continue
		}

		if funk.ContainsString(dependenyChain, childInsDef.Operator) {
			return nil, fmt.Errorf("recursion in %s", childInsDef.Name)
		}

		insOpId, err := uuid.Parse(childInsDef.Operator)

		if err != nil {
			return opDef, fmt.Errorf(`id is not a valid UUID v4: "%s" --> "%s"`, opDef.Id, err)
		}

		childOpDef, err := s.load(insOpId, dependenyChain)

		if err != nil {
			continue
		}

		childInsDef.OperatorDef = *childOpDef
	}

	return opDef, nil
}
