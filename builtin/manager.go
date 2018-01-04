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

func MakeOperator(def op.InstanceDef, par *op.Operator) (*op.Operator, error) {
	if creator := getCreatorFunc(def.Operator); creator != nil {
		return creator(def, par)
	}
	return nil, errors.New("unknown builtin operator")
}

func Register(name string, creator CreatorFunc) {
	m.creators[name] = creator
}

func init() {
	m.creators = make(map[string]CreatorFunc)
	Register("function", functionCreator)
}

func getCreatorFunc(name string) CreatorFunc {
	c, _ := m.creators[name]
	return c
}
