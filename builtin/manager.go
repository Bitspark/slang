package builtin

import (
	"errors"
	"slang/op"
)

type CreatorFunc func(op.InstanceDef, *op.Operator) (*op.Operator, error)

type Manager struct {
	creators map[string]CreatorFunc
}

var m = Manager{}

func init() {
	m.creators = make(map[string]CreatorFunc)
	Register("function", functionCreator)
}

func MakeOperator(def op.InstanceDef, par *op.Operator) (*op.Operator, error) {
	creator, ok := m.creators[def.Operator]

	if !ok {
		return nil, errors.New("unknown builtin operator")
	}

	return creator(def, par)
}

func Register(name string, creator CreatorFunc) {
	m.creators[name] = creator
}
