package elem

import (
	"testing"

	"github.com/Bitspark/slang/tests/assertions"
)

func Test_TimeDelay__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocDelay := getBuiltinCfg(timeDelayId)
	a.NotNil(ocDelay)
}
