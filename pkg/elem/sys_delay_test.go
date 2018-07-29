package elem

import (
	"testing"
	"github.com/Bitspark/slang/tests/assertions"
)

func TestOperatorDelay__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocDelay := getBuiltinCfg("slang.time.delay")
	a.NotNil(ocDelay)
}
