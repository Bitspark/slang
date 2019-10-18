package elem

import (
	"errors"
	"sync"

	"github.com/Bitspark/go-funk"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

type builtinConfig struct {
	opConnFunc core.CFunc
	opFunc     core.OFunc
	blueprint  core.Blueprint
}

var cfgs map[uuid.UUID]*builtinConfig
var name2Id map[string]uuid.UUID

func MakeOperator(def core.InstanceDef) (*core.Operator, error) {
	cfg := getBuiltinCfg(def.Operator)

	if cfg == nil {
		return nil, errors.New("unknown builtin operator")
	}

	if err := def.Blueprint.GenericsSpecified(); err != nil {
		return nil, err
	}

	o, err := core.NewOperator(def.Name, cfg.opFunc, cfg.opConnFunc, def.Generics, def.Properties, def.Blueprint)
	if err != nil {
		return nil, err
	}

	return o, nil
}

func GetBlueprint(id uuid.UUID) (*core.Blueprint, error) {
	cfg, ok := cfgs[id]

	if !ok {
		return nil, errors.New("builtin operator not found")
	}

	blueprint := cfg.blueprint.Copy(true)
	return &blueprint, nil
}

func IsRegistered(id uuid.UUID) bool {
	_, b := cfgs[id]
	return b
}

func Register(cfg *builtinConfig) {
	cfg.blueprint.Elementary = cfg.blueprint.Id

	id := cfg.blueprint.Id
	cfgs[id] = cfg
	name2Id[cfg.blueprint.Meta.Name] = id
}

func GetBuiltinIds() []uuid.UUID {
	return funk.Keys(cfgs).([]uuid.UUID)
}

func init() {
	cfgs = make(map[uuid.UUID]*builtinConfig)
	name2Id = make(map[string]uuid.UUID)

	Register(metaStoreCfg)

	// Data manipulating operators
	Register(dataValueCfg)
	Register(dataEvaluateCfg)
	Register(dataConvertCfg)
	Register(dataUUIDCfg)
	Register(dataVariableSetCfg)
	Register(dataVariableGetCfg)

	// Flow control operators
	Register(controlSplitCfg)
	Register(controlSwitchCfg)
	Register(controlTakeCfg)
	Register(controlLoopCfg)
	Register(controlIterateCfg)
	Register(controlReduceCfg)
	Register(controlSemaphorePCfg)
	Register(controlSemaphoreVCfg)

	// Stream accessing and processing operators
	Register(streamSerializeCfg)
	Register(streamParallelizeCfg)
	Register(streamConcatenateCfg)
	Register(streamMapAccessCfg)
	Register(streamWindowCfg)
	Register(streamWindowCollectCfg)
	Register(streamWindowReleaseCfg)
	Register(streamMapToStreamCfg)
	Register(streamStreamToMapCfg)
	Register(streamSliceCfg)
	Register(streamTransformCfg)
	Register(streamDistinctCfg)

	// Miscellaneous operators
	Register(netHTTPServerCfg)
	Register(netHTTPClientCfg)
	Register(netSendEmailCfg)
	Register(netMQTTPublishCfg)
	Register(netMQTTSubscribeCfg)

	Register(filesReadCfg)
	Register(filesWriteCfg)
	Register(filesAppendCfg)
	Register(filesReadLinesCfg)
	Register(filesZIPPackCfg)
	Register(filesZIPUnpackCfg)

	Register(encodingCSVReadCfg)
	Register(encodingCSVWriteCfg)
	Register(encodingJSONReadCfg)
	Register(encodingJSONWriteCfg)
	Register(encodingXLSXReadCfg)
	Register(encodingURLWriteCfg)

	Register(timeDelayCfg)
	Register(timeCrontabCfg)
	Register(timeParseDateCfg)
	Register(timeDateNowCfg)
	Register(timeUNIXMillisCfg)

	Register(stringTemplateCfg)
	Register(stringFormatCfg)
	Register(stringSplitCfg)
	Register(stringBeginswithCfg)
	Register(stringContainsCfg)
	Register(stringEndswithCfg)

	Register(databaseQueryCfg)
	Register(databaseExecuteCfg)
	Register(databaseKafkaSubscribeCfg)
	Register(databaseRedisGetCfg)
	Register(databaseRedisSetCfg)
	Register(databaseRedisHGetCfg)
	Register(databaseRedisHSetCfg)
	Register(databaseRedisLPushCfg)
	Register(databaseRedisHIncrByCfg)
	Register(databaseRedisSubscribeCfg)
	Register(databaseMemoryReadCfg)
	Register(databaseMemoryWriteCfg)

	Register(imageDecodeCfg)
	Register(imageEncodeCfg)

	Register(shellExecuteCfg)
	Register(systemLogCfg)

	variableStores = make(map[string]*variableStore)
	variableMutex = &sync.Mutex{}

	windowStores = make(map[string]*windowStore)
	windowMutex = &sync.Mutex{}

	memoryStores = make(map[string]*memoryStore)
	memoryMutex = &sync.Mutex{}

	semaphoreStores = make(map[string]*semaphoreStore)
	semaphoreMutex = &sync.Mutex{}
}

func getBuiltinCfg(id uuid.UUID) *builtinConfig {
	c, _ := cfgs[id]
	return c
}

// Mainly for testing

func buildOperator(insDef core.InstanceDef) (*core.Operator, error) {
	blueprint, err := GetBlueprint(insDef.Operator)

	if err != nil {
		return nil, err
	}

	if err = blueprint.SpecifyOperator(insDef.Generics, insDef.Properties); err != nil {
		return nil, err
	}
	insDef.Blueprint = *blueprint

	return MakeOperator(insDef)
}
