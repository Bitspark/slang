package elem

import (
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/Bitspark/slang/pkg/core"
	"time"
	"github.com/Bitspark/slang/tests/assertions"
)

func Test_StreamWindowTriggered__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocWnd := getBuiltinCfg("slang.stream.WindowTriggered")
	a.NotNil(ocWnd)
}

func Test_StreamWindowTriggered(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)
	o, err := buildOperator(
		core.InstanceDef{
			Name:     "testop",
			Operator: "slang.stream.WindowTriggered",
			Generics: map[string]*core.TypeDef{
				"itemType": {
					Type: "string",
				},
			},
		},
	)
	r.NotNil(o)
	r.NoError(err)
	o.Main().Out().Bufferize()
	o.Start()

	p := func() {
		<-time.After(10 * time.Millisecond)
	}

	p()
	o.Main().In().Map("trigger").PushBOS()
	p()
	o.Main().In().Map("stream").Push([]interface{}{1, 2})
	p()
	o.Main().In().Map("trigger").Stream().Push(1)
	p()
	o.Main().In().Map("stream").Push([]interface{}{3})
	p()
	o.Main().In().Map("stream").Push([]interface{}{4, 5})
	p()
	o.Main().In().Map("trigger").Stream().Push(1)
	p()
	o.Main().In().Map("stream").Push([]interface{}{6, 7})
	p()
	o.Main().In().Map("stream").Push([]interface{}{8})
	p()
	o.Main().In().Map("trigger").Stream().Push(1)
	p()
	o.Main().In().Map("trigger").PushEOS()
	p()
	o.Main().In().Map("stream").Push([]interface{}{9, 10, 11})
	p()
	o.Main().In().Map("trigger").PushBOS()
	p()
	o.Main().In().Map("trigger").PushEOS()
	p()
	o.Main().In().Map("stream").Push([]interface{}{9, 10, 11})
	p()
	o.Main().In().Map("trigger").PushBOS()
	p()
	o.Main().In().Map("stream").Push([]interface{}{12, 13, 14})
	p()
	o.Main().In().Map("trigger").Stream().Push(1)
	p()
	o.Main().In().Map("trigger").PushEOS()
	p()

	o.Main().Out().PullBOS()
	a.PortPushes([]interface{}{1, 2}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{3, 4, 5}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{6, 7, 8}, o.Main().Out().Stream())
	o.Main().Out().PullEOS()
	o.Main().Out().PullBOS()
	o.Main().Out().PullEOS()
	o.Main().Out().PullBOS()
	a.PortPushes([]interface{}{12, 13, 14}, o.Main().Out().Stream())
	o.Main().Out().PullEOS()
}
