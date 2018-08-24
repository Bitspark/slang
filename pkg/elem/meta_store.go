package elem

import (
	"github.com/Bitspark/slang/pkg/core"
)

// attachStore attaches an interface array to the port and starts one or multiple go routine for this port which listen
// at the port
func attachStore(p *core.Port, key string, store map[string][]interface{}) {
	if p.Primitive() {
		store[key] = make([]interface{}, 0)
		go func() {
			for !p.Operator().Stopped() {
				store[key] = append(store[key], p.Pull())
			}
		}()
	} else if p.Type() == core.TYPE_MAP {
		for _, sub := range p.MapEntries() {
			attachStore(p.Map(sub), key + "." + sub, store)
		}
	} else if p.Type() == core.TYPE_STREAM {
		attachStore(p.Stream(), key, store)
	}
}

var metaStoreCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "generic",
					Generic: "examineType",
				},
				Out: core.TypeDef{
					Type: "trigger",
				},
			},
			"query": {
				In: core.TypeDef{
					Type: "trigger",
				},
				Out: core.TypeDef{
					Type: "generic",
					Generic: "examineType",
				},
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		querySrv := op.Service("query")
		queryIn := querySrv.In()
		queryOut := querySrv.Out()

		store := make(map[string][]interface{})
		// starts listening go routines
		attachStore(in, "in", store)

		for !op.CheckStop() {
			queryIn.Pull()
			queryOut.Push()
		}
	},
}
