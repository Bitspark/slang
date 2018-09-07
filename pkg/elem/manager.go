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

	Register("slang.meta.Store", metaStoreCfg)

	// Data manipulating operators
	Register("slang.data.Value", dataValueCfg)
	Register("slang.data.Evaluate", dataEvaluateCfg)
	Register("slang.data.Convert", dataConvertCfg)

	// Flow control operators
	Register("slang.control.Split", controlSplitCfg)
	Register("slang.control.Merge", controlMergeCfg)
	Register("slang.control.Switch", controlSwitchCfg)
	// Register("slang.control.SingleSplit", controlSingleSplitCfg)
	Register("slang.control.Choose", controlChooseCfg)
	Register("slang.control.Take", controlTakeCfg)
	Register("slang.control.Loop", controlLoopCfg)
	Register("slang.control.Iterate", controlIterateCfg)
	Register("slang.control.Reduce", controlReduceCfg)

	// Stream accessing and processing operators
	Register("slang.stream.Serialize", streamSerializeCfg)
	Register("slang.stream.Parallelize", streamParallelizeCfg)
	Register("slang.stream.Concatenate", streamConcatenateCfg)
	Register("slang.stream.MapAccess", streamMapAccessCfg)
	Register("slang.stream.WindowCount", streamWindowCountCfg)
	Register("slang.stream.WindowTriggered", streamWindowTriggeredCfg)

	// Miscellaneous operators
	Register("slang.net.HTTPServer", netHTTPServerCfg)
	Register("slang.net.HTTPClient", netHTTPClientCfg)
	Register("slang.net.SendEmail", netSendEmailCfg)
	Register("slang.net.MQTTPublish", netMQTTPublishCfg)
	Register("slang.net.MQTTSubscribe", netMQTTSubscribeCfg)

	Register("slang.files.Read", filesReadCfg)
	Register("slang.files.Write", filesWriteCfg)
	Register("slang.files.ZIPPack", filesZIPPackCfg)
	Register("slang.files.ZIPUnpack", filesZIPUnpackCfg)

	Register("slang.encoding.CSVRead", encodingCSVReadCfg)
	Register("slang.encoding.JSONRead", encodingJSONReadCfg)
	Register("slang.encoding.JSONWrite", encodingJSONWriteCfg)
	Register("slang.encoding.XLSXRead", encodingXLSXReadCfg)
	Register("slang.encoding.URLWrite", encodingURLWriteCfg)

	Register("slang.time.Delay", timeDelayCfg)

	Register("slang.string.Template", stringTemplateCfg)
	Register("slang.string.Format", stringFormatCfg)
	Register("slang.string.Split", stringSplitCfg)

	Register("slang.database.Query", databaseQueryCfg)
	Register("slang.database.Execute", databaseExecuteCfg)

	Register("slang.image.Decode", imageDecodeCfg)
	Register("slang.image.Encode", imageEncodeCfg)

	Register("slang.system.Execute", systemExecuteCfg)
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