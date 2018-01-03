package builtin

import (
	"slang"
	"errors"
)

type CreatorFunc func(slang.InstanceDef) (*slang.Operator, error)

type Manager struct {
	creators map[string]CreatorFunc
}

var m = Manager{}

func init() {
	m.creators = make(map[string]CreatorFunc)
	m.Register("function", functionCreator)
}

func M() Manager {
	return m
}

func (m Manager) MakeOperator(def slang.InstanceDef) (*slang.Operator, error) {
	creator, ok := m.creators[def.Operator]

	if !ok {
		return nil, errors.New("unknown builtin operator")
	}

	return creator(def)
}

func (m Manager) Register(name string, creator CreatorFunc) {
	m.creators[name] = creator
}
