package storage

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

type Storage interface {
	IsLibrary(opID uuid.UUID) bool
	List() ([]uuid.UUID, error)

	Load(opId uuid.UUID) (*core.OperatorDef, error)
	Store(opDef core.OperatorDef) (uuid.UUID, error)
}
