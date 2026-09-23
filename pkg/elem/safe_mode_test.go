package elem

import "testing"

func TestSafeModeLeavesHostEffectsUnregistered(t *testing.T) {
	SafeMode = true
	Init()
	t.Cleanup(func() {
		SafeMode = false
		Init()
	})
	for _, cfg := range []*builtinConfig{shellExecuteCfg, filesWriteCfg, filesAppendCfg, filesZIPPackCfg} {
		if IsRegistered(cfg.blueprint.Id) {
			t.Errorf("safe mode registered %s", cfg.blueprint.Meta.Name)
		}
	}
	if !IsRegistered(netHTTPClientCfg.blueprint.Id) || !IsRegistered(dataEvaluateCfg.blueprint.Id) {
		t.Error("safe mode must keep computation and HTTP")
	}
}
