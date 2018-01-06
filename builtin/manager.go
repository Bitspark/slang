package builtin

import (
	"errors"
	"slang/core"
)

type CreatorFunc func(core.InstanceDef, *core.Operator) (*core.Operator, error)

type Manager struct {
	creators map[string]CreatorFunc
}

var m = Manager{}

func MakeOperator(def core.InstanceDef, par *core.Operator) (*core.Operator, error) {
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
	Register("eval", createOpEval)
	Register("fork", createOpFork)
	Register("loop", createOpLoop)
	Register("merge", createOpMerge)
}

func getCreatorFunc(name string) CreatorFunc {
	c, _ := m.creators[name]
	return c
}
