package builtin

import (
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/Bitspark/slang/pkg/core"
	"time"
	"github.com/Bitspark/slang/tests/assertions"
)

func TestOperatorWindowTriggered__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocWnd := getBuiltinCfg("slang.window.triggered")
	a.NotNil(ocWnd)
}

func TestOperatorWindowTriggered(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)
	o, err := MakeOperator(
		core.InstanceDef{
			Name:     "testop",
			Operator: "slang.window.triggered",
			Generics: map[string]*core.PortDef{
				"itemType": {
					Type: "string",
				},
			},
		},
	)
	r.NotNil(o)
	r.NoError(err)
	o.Out().Bufferize()
	o.Start()

	p := func() {
		<-time.After(10 * time.Millisecond)
	}

	p()
	o.In().Map("trigger").PushBOS()
	p()
	o.In().Map("stream").Push([]interface{}{1, 2})
	p()
	o.In().Map("trigger").Stream().Push(1)
	p()
	o.In().Map("stream").Push([]interface{}{3})
	p()
	o.In().Map("stream").Push([]interface{}{4, 5})
	p()
	o.In().Map("trigger").Stream().Push(1)
	p()
	o.In().Map("stream").Push([]interface{}{6, 7})
	p()
	o.In().Map("stream").Push([]interface{}{8})
	p()
	o.In().Map("trigger").Stream().Push(1)
	p()
	o.In().Map("trigger").PushEOS()
	p()
	o.In().Map("stream").Push([]interface{}{9, 10, 11})
	p()
	o.In().Map("trigger").PushBOS()
	p()
	o.In().Map("trigger").PushEOS()
	p()
	o.In().Map("stream").Push([]interface{}{9, 10, 11})
	p()
	o.In().Map("trigger").PushBOS()
	p()
	o.In().Map("stream").Push([]interface{}{12, 13, 14})
	p()
	o.In().Map("trigger").Stream().Push(1)
	p()
	o.In().Map("trigger").PushEOS()
	p()

	o.Out().PullBOS()
	a.PortPushes([]interface{}{1, 2}, o.Out().Stream())
	a.PortPushes([]interface{}{3, 4, 5}, o.Out().Stream())
	a.PortPushes([]interface{}{6, 7, 8}, o.Out().Stream())
	o.Out().PullEOS()
	o.Out().PullBOS()
	o.Out().PullEOS()
	o.Out().PullBOS()
	a.PortPushes([]interface{}{12, 13, 14}, o.Out().Stream())
	o.Out().PullEOS()
}
