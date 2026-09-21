package elem

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	Init()
	os.Exit(m.Run())
}

// Retired operators are exercised without adding them back to the public registry.
func registerLegacyOperator(t *testing.T, cfg *builtinConfig) {
	t.Helper()
	Init()
	Register(cfg)
	t.Cleanup(Init)
}
