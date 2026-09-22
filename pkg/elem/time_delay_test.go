package elem

import (
	"testing"
	"time"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/stretchr/testify/require"
)

func Test_TimeDelay__IsRegistered(t *testing.T) {
	Init()
	a := assertions.New(t)

	ocDelay := getBuiltinCfg(timeDelayId)
	a.NotNil(ocDelay)
}

func Test_TimeDelay__DelaysAndPreservesItems(t *testing.T) {
	Init()
	o, err := buildOperator(core.InstanceDef{
		Operator: timeDelayId,
		Generics: core.Generics{"itemType": {Type: "string"}},
	})
	require.NoError(t, err)
	o.Main().Out().Bufferize()
	o.Start()
	defer o.Stop()

	for _, delay := range []time.Duration{40 * time.Millisecond, 0} {
		started := time.Now()
		o.Main().In().Push(map[string]interface{}{
			"item": "unchanged", "delay": float64(delay / time.Millisecond),
		})
		result := make(chan interface{}, 1)
		go func() { result <- o.Main().Out().Pull() }()
		select {
		case item := <-result:
			require.Equal(t, "unchanged", item)
			require.GreaterOrEqual(t, time.Since(started), delay)
		case <-time.After(2 * time.Second):
			t.Fatal("delay operator did not emit the item")
		}
	}
}
