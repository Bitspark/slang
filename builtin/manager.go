package builtin

import (
	"slang"
	"errors"
)

type CreatorFunc func(properties map[string]interface{}) (*slang.Operator, error)

type Manager struct {
	creators map[string]CreatorFunc
}

var m = Manager{}

func init() {
	m.Register("function", functionCreator)
}

func M() Manager {
	return m
}

func (m Manager) MakeOperator(name string, properties map[string]interface{}) (*slang.Operator, error) {
	creator, ok := m.creators[name]

	if !ok {
		return nil, errors.New("unknown builtin operator")
	}

	return creator(properties)
}

func (m Manager) Register(name string, creator CreatorFunc) {
	m.creators[name] = creator
}
