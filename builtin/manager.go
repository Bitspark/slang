package builtin

import (
	"errors"
	"slang/core"
)

type CreatorFunc func(core.InstanceDef) (*core.Operator, error)
type PropertyFunc func(*core.Operator, map[string]interface{}) error

type builtinConfig struct {
	oPropFunc PropertyFunc
	oFunc     core.OFunc
	oDef      *core.OperatorDef
}

var cfgs map[string]*builtinConfig

func MakeOperator(def core.InstanceDef) (*core.Operator, error) {
	cfg := getBuiltinCfg(def.Operator)

	if cfg == nil {
		return nil, errors.New("unknown builtin operator")
	}

	if def.In == nil || def.Out == nil {
		return nil, errors.New("port definitions missing")
	}

	// TODO: use this:
	// o, err := core.NewOperator(def.Name, cfg.ofunc, *cfg.odef.In, *cfg.odef.Out)
	o, err := core.NewOperator(def.Name, cfg.oFunc, *def.In, *def.Out)
	if err != nil {
		return nil, err
	}

	if cfg.oPropFunc != nil {
		err = cfg.oPropFunc(o, def.Properties)
		if err != nil {
			return nil, err
		}
	}

	return o, nil
}

func GetOperatorDef(name string) *core.OperatorDef {
	cfg, _ := cfgs[name]
	return cfg.oDef
}

func IsRegistered(name string) bool {
	_, b := cfgs[name]
	return b
}

func Register(name string, cfg *builtinConfig) {
	cfgs[name] = cfg
}

func init() {
	cfgs = make(map[string]*builtinConfig)
	Register("eval", evalOpCfg)
	Register("fork", forkOpCfg)
	Register("loop", loopOpCfg)
	Register("merge", mergeOpCfg)
}

func getBuiltinCfg(name string) *builtinConfig {
	c, _ := cfgs[name]
	return c
}
