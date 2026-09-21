package tests

import (
	"os"
	"testing"

	"github.com/Bitspark/slang/pkg/elem"
)

func TestMain(m *testing.M) {
	elem.Init()
	tl.Reload()
	os.Exit(m.Run())
}
