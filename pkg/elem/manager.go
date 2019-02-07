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

func Register(idstr string, name string, cfg *builtinConfig) {
	cfg.opDef.Id = idstr
	cfg.opDef.Meta.Name = name
	cfg.opDef.Elementary = idstr

	id := getId(idstr)
	cfgs[id] = cfg
	name2Id[name] = id
}

func GetBuiltinIds() []uuid.UUID {
	return funk.Keys(cfgs).([]uuid.UUID)
}

func init() {
	cfgs = make(map[uuid.UUID]*builtinConfig)
	name2Id = make(map[string]uuid.UUID)

	Register("cf20bcec-2028-45b4-a00c-0ce348c381c4", "slang.meta.Store", metaStoreCfg)

	// Data manipulating operators
	Register("8b62495a-e482-4a3e-8020-0ab8a350ad2d", "slang.data.Value", dataValueCfg)
	Register("37ccdc28-67b0-4bb1-8591-4e0e813e3ec1", "slang.data.Evaluate", dataEvaluateCfg)
	Register("d1191456-3583-4eaf-8ec1-e486c3818c60", "slang.data.Convert", dataConvertCfg)

	// Flow control operators
	Register("fed72b41-2584-424c-8213-1978410ccab6", "slang.control.Split", controlSplitCfg)
	Register("97583526-178b-42ca-b73c-9491ed8536f2", "slang.control.Merge", controlMergeCfg)
	Register("cd6fc5c8-5b64-4b1a-9885-59ede141b398", "slang.control.Switch", controlSwitchCfg)
	Register("9bebc4bf-d512-4944-bcb1-5b2c3d5b5471", "slang.control.Take", controlTakeCfg)
	Register("0b8a1592-1368-44bc-92d5-692acc78b1d3", "slang.control.Loop", controlLoopCfg)
	Register("e58624d4-5568-40d3-8b77-ab792ef620f1", "slang.control.Iterate", controlIterateCfg)
	Register("b95e6da8-9770-4a04-a73d-cdfe2081870f", "slang.control.Reduce", controlReduceCfg)

	// Stream accessing and processing operators
	Register("13257172-b05d-497c-be23-da7c86577c1e", "slang.stream.Serialize", streamSerializeCfg)
	Register("b8428777-7667-4012-b76a-a5b7f4d1e433", "slang.stream.Parallelize", streamParallelizeCfg)
	Register("fb174c53-80bd-4e29-955a-aafe33ebfb30", "slang.stream.Concatenate", streamConcatenateCfg)
	Register("618c4007-70fc-44ac-9443-184df77ab730", "slang.stream.MapAccess", streamMapAccessCfg)
	Register("5b704038-9617-454a-b7a1-2091277cff69", "slang.stream.Window", streamWindowCfg)
	Register("14f5de1a-5e38-4f9c-a625-eff7a572078c", "slang.stream.WindowCollect", streamWindowCollectCfg)
	Register("47b3f097-2043-42c6-aad5-0cfdb9004aef", "slang.stream.WindowRelease", streamWindowReleaseCfg)
	Register("d099a1cd-69eb-43a2-b95b-239612c457fc", "slang.stream.MapToStream", streamMapToStreamCfg)
	Register("42d0f961-4ce0-4a20-b1b0-3da46396ae66", "slang.stream.StreamToMap", streamStreamToMapCfg)
	Register("2471a7aa-c5b9-4392-b23f-d0c7bcdb3f39", "slang.stream.Slice", streamSliceCfg)
	Register("dce082cb-7272-4e85-b4fa-740778e8ba8d", "slang.stream.Transform", streamTransformCfg)

	// Miscellaneous operators
	Register("241cc7ef-c6d6-49c1-8729-c5e3c0be8188", "slang.net.HTTPServer", netHTTPServerCfg)
	Register("f7f5907d-758b-4892-8a3e-ae86b877b869", "slang.net.HTTPClient", netHTTPClientCfg)
	Register("741b8a21-0b6d-40e5-a281-b179a49e9030", "slang.net.SendEmail", netSendEmailCfg)
	Register("c6b5bef6-e93e-4bc1-8ded-49c90919f39d", "slang.net.MQTTPublish", netMQTTPublishCfg)
	Register("fd51e295-3483-4558-9b26-8c16d579c4ef", "slang.net.MQTTSubscribe", netMQTTSubscribeCfg)

	Register("f7eecf2c-6504-478f-b2fa-809bec71463c", "slang.files.Read", filesReadCfg)
	Register("9b61597d-cfbc-42d1-9620-210081244ba1", "slang.files.Write", filesWriteCfg)
	Register("e49369c2-eac2-4dc7-9a6d-b635ae1654f9", "slang.files.Append", filesAppendCfg)
	Register("6124cd6b-5c23-4e17-a714-458d0f8ac1a7", "slang.files.ReadLines", filesReadLinesCfg)
	Register("dc5325bc-a816-47c8-8a8a-f741497459f7", "slang.files.ZIPPack", filesZIPPackCfg)
	Register("04714d4a-1d5d-4b68-b614-524dd4662ef4", "slang.files.ZIPUnpack", filesZIPUnpackCfg)

	Register("77d60459-f8b5-4f4b-b293-740164c49a82", "slang.encoding.CSVRead", encodingCSVReadCfg)
	Register("b79b019f-5efe-4012-9a1d-1f61549ede25", "slang.encoding.JSONRead", encodingJSONReadCfg)
	Register("d4aabe2d-dee7-409f-b2bb-713ebc836672", "slang.encoding.JSONWrite", encodingJSONWriteCfg)
	Register("69db81cf-2a24-4470-863f-ceffaeb8b246", "slang.encoding.XLSXRead", encodingXLSXReadCfg)
	Register("702a2036-a1cc-4783-8b83-b18494c5e9f1", "slang.encoding.URLWrite", encodingURLWriteCfg)

	Register("7d61b83a-9aa2-4875-9c21-1e11f6adbfae", "slang.time.Delay", timeDelayCfg)
	Register("60b849fd-ca5a-4206-8312-996e4e3f6c31", "slang.time.Crontab", timeCrontabCfg)
	Register("2a9da2d5-2684-4d2f-8a37-9560d0f2de29", "slang.time.ParseDate", timeParseDateCfg)
	Register("808c7846-db9f-43ee-989b-37a08ce7e70d", "slang.time.Now", timeDateNowCfg)

	Register("3c39f999-b5c2-490d-aed1-19149d228b04", "slang.string.Template", stringTemplateCfg)
	Register("21dbddf2-2d07-494e-8950-3ac0224a3ff5", "slang.string.Format", stringFormatCfg)
	Register("c02bc7ad-65e5-4a43-a2a3-7d86b109915d", "slang.string.Split", stringSplitCfg)
	Register("9f274995-2726-4513-ac7c-f15ac7b68720", "slang.string.StartsWith", stringBeginswithCfg)
	Register("8a01dfe3-5dcf-4f40-9e54-f5b168148d2a", "slang.string.Contains", stringContainsCfg)
	Register("db8b1677-baaf-4072-8047-0359cd68be9e", "slang.string.Endswith", stringEndswithCfg)

	Register("ce3a3e0e-d579-4712-8573-713a645c2271", "slang.database.Query", databaseQueryCfg)
	Register("e5abeb01-3aad-47f3-a753-789a9fff0d50", "slang.database.Execute", databaseExecuteCfg)

	Register("4b082c52-9a99-472f-9277-f5ca9651dbfb", "slang.image.Decode", imageDecodeCfg)
	Register("bd4475af-795b-4be8-9e57-9fec9444e028", "slang.image.Encode", imageEncodeCfg)

	Register("13cbad40-da00-40d7-bdcd-981b14ec346b", "slang.system.Execute", systemExecuteCfg)

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
