package elem

import (
	"errors"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/go-funk"
)

type builtinConfig struct {
	opConnFunc core.CFunc
	opFunc     core.OFunc
	opDef      core.OperatorDef
}

var cfgs map[string]*builtinConfig

func MakeOperator(def core.InstanceDef) (*core.Operator, error) {
	cfg := getBuiltinCfg(def.Operator)

	if cfg == nil {
		return nil, errors.New("unknown builtin operator")
	}
	
	if err := def.OperatorDef.GenericsSpecified(); err != nil {
		return nil, err
	}

	o, err := core.NewOperator(def.Name, cfg.opFunc, cfg.opConnFunc, def.Generics, def.Properties, def.OperatorDef)
	if err != nil {
		return nil, err
	}

	return o, nil
}

func GetOperatorDef(operator string) (core.OperatorDef, error) {
	cfg, ok := cfgs[operator]
	if !ok {
		return core.OperatorDef{}, errors.New("builtin operator not found")
	}

	return cfg.opDef.Copy(), nil
}

func IsRegistered(name string) bool {
	_, b := cfgs[name]
	return b
}

func Register(name string, cfg *builtinConfig) {
	cfgs[name] = cfg
	cfg.opDef.Elementary = name
}

func GetBuiltinNames() []string {
	return funk.Keys(cfgs).([]string)
}

func init() {
	cfgs = make(map[string]*builtinConfig)

	// Data manipulating operators
	Register("slang.data.Constant", constOpCfg)
	Register("slang.data.Evaluate", evalOpCfg)
	Register("slang.data.Convert", convertOpCfg)

	// Flow control operators
	Register("slang.control.Split", splitOpCfg)
	Register("slang.control.Merge", mergeOpCfg)
	Register("slang.control.Switch", switchOpCfg)
	Register("slang.control.SingleSplit", singleSplitOpCfg)
	Register("slang.control.SingleMerge", singleMergeOpCfg)

	// Stream producing and consuming operators
	Register("slang.stream.Loop", loopOpCfg)
	Register("slang.stream.Iterate", iterateOpCfg)
	Register("slang.stream.Reduce", reduceOpCfg)

	Register("slang.deprecated.take", takeOpCfg) // TODO: Add functionality to Merge operator

	Register("slang.stream.extract", extractOpCfg)
	Register("slang.stream.concat", concatOpCfg)
	Register("slang.stream.serialize", serializeOpCfg)
	Register("slang.stream.mapAccess", mapAccessOpCfg)

	Register("slang.window.count", windowCountOpCfg)
	Register("slang.window.triggered", windowTriggeredOpCfg)

	Register("slang.net.httpServer", httpServerOpCfg)
	Register("slang.net.sendEmail", sendEmailOpCfg)

	Register("slang.files.read", fileReadOpCfg)
	Register("slang.files.write", fileWriteOpCfg)

	Register("slang.encoding.csv.read", csvReadOpCfg)

	Register("slang.encoding.json.write", jsonWriteOpCfg)
	Register("slang.encoding.json.read", jsonReadOpCfg)

	Register("slang.encoding.xlsx.read", xlsxReadOpCfg)

	Register("slang.time.delay", delayOpCfg)

	Register("slang.template.format", templateFormatOpCfg)
}

func getBuiltinCfg(name string) *builtinConfig {
	c, _ := cfgs[name]
	return c
}

// Mainly for testing

func buildOperator(insDef core.InstanceDef) (*core.Operator, error) {
	insDef.OperatorDef, _ = GetOperatorDef(insDef.Operator)
	err := insDef.OperatorDef.SpecifyOperator(insDef.Generics, insDef.Properties)
	if err != nil {
		return nil, err
	}
	return MakeOperator(insDef)
}