package elem

import (
	"errors"
	"github.com/Bitspark/go-funk"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
	"sync"
)

type builtinConfig struct {
	opConnFunc core.CFunc
	opFunc     core.OFunc
	opDef      core.OperatorDef
}

var cfgs map[uuid.UUID]*builtinConfig
var name2Id map[string]uuid.UUID

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

func getId(idOrName string) uuid.UUID {
	if id, ok := name2Id[idOrName]; ok {
		return id
	}
	id, _ := uuid.Parse(idOrName)
	return id
}

func GetOperatorDef(idOrName string) (*core.OperatorDef, error) {
	cfg, ok := cfgs[getId(idOrName)]

	if !ok {
		return nil, errors.New("builtin operator not found")
	}

	opDef := cfg.opDef.Copy(true)
	return &opDef, nil
}

func IsRegistered(idOrName string) bool {
	_, b := cfgs[getId(idOrName)]
	return b
}

func Register(cfg *builtinConfig) {
	cfg.opDef.Elementary = cfg.opDef.Id

	id := getId(cfg.opDef.Id)
	cfgs[id] = cfg
	name2Id[cfg.opDef.Meta.Name] = id
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

	// Flow control operators
	Register(controlSplitCfg)
	Register(controlMergeCfg)
	Register(controlSwitchCfg)
	Register(controlTakeCfg)
	Register(controlLoopCfg)
	Register(controlIterateCfg)
	Register(controlReduceCfg)

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

	// Miscellaneous operators
	Register(netHTTPServerCfg)
	Register(netHTTPClientCfg)
	//Register("741b8a21-0b6d-40e5-a281-b179a49e9030", "slang.net.SendEmail", netSendEmailCfg)
	//Register("c6b5bef6-e93e-4bc1-8ded-49c90919f39d", "slang.net.MQTTPublish", netMQTTPublishCfg)
	//Register("fd51e295-3483-4558-9b26-8c16d579c4ef", "slang.net.MQTTSubscribe", netMQTTSubscribeCfg)
	//
	//Register("f7eecf2c-6504-478f-b2fa-809bec71463c", "slang.files.Read", filesReadCfg)
	//Register("9b61597d-cfbc-42d1-9620-210081244ba1", "slang.files.Write", filesWriteCfg)
	//Register("e49369c2-eac2-4dc7-9a6d-b635ae1654f9", "slang.files.Append", filesAppendCfg)
	//Register("6124cd6b-5c23-4e17-a714-458d0f8ac1a7", "slang.files.ReadLines", filesReadLinesCfg)
	//Register("dc5325bc-a816-47c8-8a8a-f741497459f7", "slang.files.ZIPPack", filesZIPPackCfg)
	//Register("04714d4a-1d5d-4b68-b614-524dd4662ef4", "slang.files.ZIPUnpack", filesZIPUnpackCfg)
	//
	//Register("77d60459-f8b5-4f4b-b293-740164c49a82", "slang.encoding.CSVRead", encodingCSVReadCfg)
	//Register("b79b019f-5efe-4012-9a1d-1f61549ede25", "slang.encoding.JSONRead", encodingJSONReadCfg)
	//Register("d4aabe2d-dee7-409f-b2bb-713ebc836672", "slang.encoding.JSONWrite", encodingJSONWriteCfg)
	//Register("69db81cf-2a24-4470-863f-ceffaeb8b246", "slang.encoding.XLSXRead", encodingXLSXReadCfg)
	//Register("702a2036-a1cc-4783-8b83-b18494c5e9f1", "slang.encoding.URLWrite", encodingURLWriteCfg)
	//
	//Register("7d61b83a-9aa2-4875-9c21-1e11f6adbfae", "slang.time.Delay", timeDelayCfg)
	//Register("60b849fd-ca5a-4206-8312-996e4e3f6c31", "slang.time.Crontab", timeCrontabCfg)
	//Register("2a9da2d5-2684-4d2f-8a37-9560d0f2de29", "slang.time.ParseDate", timeParseDateCfg)
	//Register("808c7846-db9f-43ee-989b-37a08ce7e70d", "slang.time.Now", timeDateNowCfg)
	//
	//Register("3c39f999-b5c2-490d-aed1-19149d228b04", "slang.string.Template", stringTemplateCfg)
	//Register("21dbddf2-2d07-494e-8950-3ac0224a3ff5", "slang.string.Format", stringFormatCfg)
	//Register("c02bc7ad-65e5-4a43-a2a3-7d86b109915d", "slang.string.Split", stringSplitCfg)
	//Register("9f274995-2726-4513-ac7c-f15ac7b68720", "slang.string.StartsWith", stringBeginswithCfg)
	//Register("8a01dfe3-5dcf-4f40-9e54-f5b168148d2a", "slang.string.Contains", stringContainsCfg)
	//Register("db8b1677-baaf-4072-8047-0359cd68be9e", "slang.string.Endswith", stringEndswithCfg)
	//
	//Register("ce3a3e0e-d579-4712-8573-713a645c2271", "slang.database.Query", databaseQueryCfg)
	//Register("e5abeb01-3aad-47f3-a753-789a9fff0d50", "slang.database.Execute", databaseExecuteCfg)
	//
	//Register("4b082c52-9a99-472f-9277-f5ca9651dbfb", "slang.image.Decode", imageDecodeCfg)
	//Register("bd4475af-795b-4be8-9e57-9fec9444e028", "slang.image.Encode", imageEncodeCfg)
	//
	//Register("13cbad40-da00-40d7-bdcd-981b14ec346b", "slang.system.Execute", systemExecuteCfg)

	windowStores = make(map[string]*windowStore)
	windowMutex = &sync.Mutex{}
}

func getBuiltinCfg(id string) *builtinConfig {
	c, _ := cfgs[getId(id)]
	return c
}

// Mainly for testing

func buildOperator(insDef core.InstanceDef) (*core.Operator, error) {
	opDef, err := GetOperatorDef(insDef.Operator)

	if err != nil {
		return nil, err
	}

	if err = opDef.SpecifyOperator(insDef.Generics, insDef.Properties); err != nil {
		return nil, err
	}

	insDef.OperatorDef = *opDef

	return MakeOperator(insDef)
}
