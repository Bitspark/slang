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
	oDef      core.OperatorDef
}

var cfgs map[string]*builtinConfig

func MakeOperator(def core.InstanceDef) (*core.Operator, error) {
	cfg := getBuiltinCfg(def.Operator)

	if cfg == nil {
		return nil, errors.New("unknown builtin operator")
	}

	in := cfg.oDef.In.Copy()
	out := cfg.oDef.Out.Copy()

	if err := in.SpecifyGenericPorts(def.Generics); err != nil {
		return nil, err
	}
	if err := out.SpecifyGenericPorts(def.Generics); err != nil {
		return nil, err
	}

	if err := in.GenericsSpecified(); err != nil {
		return nil, err
	}
	if err := out.GenericsSpecified(); err != nil {
		return nil, err
	}

	o, err := core.NewOperator(def.Name, cfg.oFunc, in, out)
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

func GetOperatorDef(insDef *core.InstanceDef) (core.OperatorDef, error) {
	cfg, ok := cfgs[insDef.Operator]
	oDef := cfg.oDef

	if !ok {
		return oDef, errors.New("builtin operator not found")
	}

	// We must not change oDef in any way as this would affect other instances of this builtin operator
	if err := oDef.SpecifyGenericPorts(insDef.Generics); err != nil {
		return oDef, err
	}

	return oDef, nil
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
	Register("const", constOpCfg)
	Register("eval", evalOpCfg)
	Register("fork", forkOpCfg)
	Register("loop", loopOpCfg)
	Register("merge", mergeOpCfg)
	Register("take", takeOpCfg)
	Register("agg", aggOpCfg)
	Register("reduce", reduceOpCfg)
	Register("syncFork", syncForkOpCfg)
	Register("syncMerge", syncMergeOpCfg)
}

func getBuiltinCfg(name string) *builtinConfig {
	c, _ := cfgs[name]
	return c
}
