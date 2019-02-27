package elem

import (
	"errors"
	"github.com/Bitspark/go-funk"
	"github.com/Bitspark/slang/pkg/core"
	"sync"
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
	Register("slang.data.UUID", dataUUIDCfg)

	// Flow control operators
	Register("slang.control.Split", controlSplitCfg)
	Register("slang.control.Merge", controlMergeCfg)
	Register("slang.control.Switch", controlSwitchCfg)
	Register("slang.control.Take", controlTakeCfg)
	Register("slang.control.Loop", controlLoopCfg)
	Register("slang.control.Iterate", controlIterateCfg)
	Register("slang.control.Reduce", controlReduceCfg)
	Register("slang.control.SemaphoreP", controlSemaphorePCfg)
	Register("slang.control.SemaphoreV", controlSemaphoreVCfg)

	// Stream accessing and processing operators
	Register("slang.stream.Serialize", streamSerializeCfg)
	Register("slang.stream.Parallelize", streamParallelizeCfg)
	Register("slang.stream.Concatenate", streamConcatenateCfg)
	Register("slang.stream.MapAccess", streamMapAccessCfg)
	Register("slang.stream.Window", streamWindowCfg)
	Register("slang.stream.WindowCollect", streamWindowCollectCfg)
	Register("slang.stream.WindowRelease", streamWindowReleaseCfg)
	Register("slang.stream.MapToStream", streamMapToStreamCfg)
	Register("slang.stream.StreamToMap", streamStreamToMapCfg)
	Register("slang.stream.Slice", streamSliceCfg)
	Register("slang.stream.Transform", streamTransformCfg)

	// Miscellaneous operators
	Register("slang.net.HTTPServer", netHTTPServerCfg)
	Register("slang.net.HTTPClient", netHTTPClientCfg)
	Register("slang.net.SendEmail", netSendEmailCfg)
	Register("slang.net.MQTTPublish", netMQTTPublishCfg)
	Register("slang.net.MQTTSubscribe", netMQTTSubscribeCfg)

	Register("slang.files.Read", filesReadCfg)
	Register("slang.files.Write", filesWriteCfg)
	Register("slang.files.Append", filesAppendCfg)
	Register("slang.files.ReadLines", filesReadLinesCfg)
	Register("slang.files.ZIPPack", filesZIPPackCfg)
	Register("slang.files.ZIPUnpack", filesZIPUnpackCfg)

	Register("slang.encoding.CSVRead", encodingCSVReadCfg)
	Register("slang.encoding.JSONRead", encodingJSONReadCfg)
	Register("slang.encoding.JSONWrite", encodingJSONWriteCfg)
	Register("slang.encoding.XLSXRead", encodingXLSXReadCfg)
	Register("slang.encoding.URLWrite", encodingURLWriteCfg)

	Register("slang.time.Delay", timeDelayCfg)
	Register("slang.time.Crontab", timeCrontabCfg)
	Register("slang.time.ParseDate", timeParseDateCfg)
	Register("slang.time.Now", timeDateNowCfg)
	Register("slang.time.UNIXMillis", timeUNIXMillisCfg)

	Register("slang.string.Template", stringTemplateCfg)
	Register("slang.string.Format", stringFormatCfg)
	Register("slang.string.Split", stringSplitCfg)
	Register("slang.string.StartsWith", stringBeginswithCfg)
	Register("slang.string.Contains", stringContainsCfg)
	Register("slang.string.Endswith", stringEndswithCfg)

	Register("slang.database.Query", databaseQueryCfg)
	Register("slang.database.Execute", databaseExecuteCfg)
	Register("slang.database.KafkaSubscribe", databaseKafjaSubscribeCfg)

	Register("slang.database.RedisGet", databaseRedisGetCfg)
	Register("slang.database.RedisSet", databaseRedisSetCfg)
	Register("slang.database.RedisHGet", databaseRedisHGetCfg)
	Register("slang.database.RedisHSet", databaseRedisHSetCfg)
	Register("slang.database.RedisLPush", databaseRedisLPushCfg)
	Register("slang.database.RedisHIncrBy", databaseRedisHIncrByCfg)

	Register("slang.database.MemoryRead", databaseMemoryReadCfg)
	Register("slang.database.MemoryWrite", databaseMemoryWriteCfg)

	Register("slang.image.Decode", imageDecodeCfg)
	Register("slang.image.Encode", imageEncodeCfg)

	Register("slang.system.Execute", systemExecuteCfg)

	windowStores = make(map[string]*windowStore)
	windowMutex = &sync.Mutex{}

	memoryStores = make(map[string]*memoryStore)
	memoryMutex = &sync.Mutex{}

	semaphoreStores = make(map[string]*semaphoreStore)
	semaphoreMutex = &sync.Mutex{}
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
