package elem

import "testing"

func TestPublicOperatorCapabilities(t *testing.T) {
	for _, cfg := range []*builtinConfig{netHTTPClientCfg, netHTTPServerCfg, netSendEmailCfg,
		netMQTTPublishCfg, netMQTTSubscribeCfg, filesReadCfg, filesReadLinesCfg, filesWriteCfg,
		filesZIPUnpackCfg, databaseQueryCfg, databaseExecuteCfg, encodingXLSXReadCfg,
		shellExecuteCfg, &builtinConfig{safe: true}} {
		if publicOperator(cfg) {
			t.Errorf("public mode permits I/O or an unknown operator: %s", cfg.blueprint.Meta.Name)
		}
	}
	if !publicOperator(dataEvaluateCfg) || !publicOperator(streamReduceCfg) {
		t.Fatal("public mode must support expressions and streams")
	}
}
