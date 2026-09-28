package elem

import (
	"fmt"
	"sync"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
	"github.com/thoas/go-funk"
)

type builtinConfig struct {
	opConnFunc core.CFunc
	opFunc     core.OFunc
	// makeFunc builds the operator function around the capabilities it needs.
	// Operators that reach the host use it instead of opFunc.
	makeFunc  func(Capabilities) (core.OFunc, error)
	blueprint core.Blueprint
	// requires declares every operation the built-in performs. A profile decides
	// from it alone whether the built-in is available (decision 0012).
	requires []Requirement
}

var Initalized bool = false

// cfgs holds every built-in this engine provides, whatever the profile.
var cfgs map[uuid.UUID]*builtinConfig

// MakeOperator builds an elementary operator that reaches this machine directly.
func MakeOperator(def core.InstanceDef) (*core.Operator, error) {
	return MakeOperatorWith(def, LocalProfile, LocalCapabilities())
}

// MakeOperatorWith builds an elementary operator that profile makes available,
// using only the given capabilities.
func MakeOperatorWith(def core.InstanceDef, profile Profile, caps Capabilities) (*core.Operator, error) {
	if err := profile.Resolve(def.Operator).Err(); err != nil {
		return nil, err
	}
	cfg := cfgs[def.Operator]

	if err := def.Blueprint.GenericsSpecified(); err != nil {
		return nil, err
	}

	opFunc := cfg.opFunc
	if cfg.makeFunc != nil {
		var err error
		if opFunc, err = cfg.makeFunc(caps); err != nil {
			return nil, fmt.Errorf("%s: %w", cfg.blueprint.Meta.Name, err)
		}
	}

	o, err := core.NewOperator(def.Name, opFunc, cfg.opConnFunc, def.Generics, def.Properties, def.Blueprint)
	if err != nil {
		return nil, err
	}

	return o, nil
}

// GetBlueprint returns a built-in's blueprint, whatever the profile, or explains
// why the ID is not a built-in this engine provides.
func GetBlueprint(id uuid.UUID) (*core.Blueprint, error) {
	cfg, ok := cfgs[id]

	if !ok {
		return nil, LocalProfile.Resolve(id).Err()
	}

	blueprint := cfg.blueprint.Copy(true)
	return &blueprint, nil
}

// IsBuiltin says whether id belongs to a built-in, including a removed one. Such
// an ID is never defined by a document.
func IsBuiltin(id uuid.UUID) bool {
	_, provided := cfgs[id]
	_, removed := removedBuiltins[id]
	return provided || removed
}

func Register(cfg *builtinConfig) {
	cfg.blueprint.Elementary = cfg.blueprint.Id
	cfgs[cfg.blueprint.Id] = cfg
}

// GetBuiltinIds lists every built-in this engine provides, whatever the profile.
func GetBuiltinIds() []uuid.UUID {
	return funk.Keys(cfgs).([]uuid.UUID)
}

func Init() {
	Initalized = true
	cfgs = make(map[uuid.UUID]*builtinConfig)

	// Data manipulating operators
	Register(dataValueCfg)
	Register(dataEvaluateCfg)
	Register(dataConvertCfg)
	Register(dataUUIDCfg)
	Register(randRangeCfg)

	// Flow control operators
	Register(controlSplitCfg)
	Register(controlMergeCfg)
	Register(controlSwitchCfg)
	Register(controlLoopCfg)
	Register(controlIterateCfg)
	Register(streamReduceCfg)
	Register(streamCtrlJoinCfg)

	// Stream accessing and processing operators
	Register(streamSerializeCfg)
	Register(streamParallelizeCfg)
	Register(streamConcatenateCfg)
	Register(streamMapAccessCfg)
	Register(streamWindow2Cfg)
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
	Register(encodingJSONPathCfg)
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
	Register(databaseMemoryReadCfg)
	Register(databaseMemoryWriteCfg)

	Register(imageDecodeCfg)
	Register(imageEncodeCfg)

	Register(systemLogCfg)

	Register(encodingPRTGHistDataCfg)

	windowStores = make(map[string]*windowStore)
	windowMutex = &sync.Mutex{}

	memoryStores = make(map[string]*memoryStore)
	memoryMutex = &sync.Mutex{}
}

func getBuiltinCfg(id uuid.UUID) *builtinConfig {
	return cfgs[id]
}

// Mainly for testing

func buildOperator(insDef core.InstanceDef) (*core.Operator, error) {
	return buildOperatorWith(insDef, LocalCapabilities())
}

func buildOperatorWith(insDef core.InstanceDef, caps Capabilities) (*core.Operator, error) {
	blueprint, err := GetBlueprint(insDef.Operator)

	if err != nil {
		return nil, err
	}

	if err = blueprint.SpecifyOperator(insDef.Generics, insDef.Properties); err != nil {
		return nil, err
	}
	insDef.Blueprint = *blueprint

	return MakeOperatorWith(insDef, LocalProfile, caps)
}
