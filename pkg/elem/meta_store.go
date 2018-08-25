package elem

import (
	"github.com/Bitspark/slang/pkg/core"
)

type storePipe struct {
	index int
	items []interface{}
}

type store map[string]*storePipe

// attachPort attaches an interface array to the port and starts one or multiple go routine for this port which listen
// at the port
func (s store) attachPort(p *core.Port, key string) {
	if p.Primitive() {
		s[key] = &storePipe{
			index: 0,
			items: []interface{}{},
		}
		go func() {
			for !p.Operator().Stopped() {
				s[key].items = append(s[key].items, p.Pull())
			}
		}()
	} else if p.Type() == core.TYPE_MAP {
		for _, sub := range p.MapEntries() {
			s.attachPort(p.Map(sub), key+"."+sub)
		}
	} else if p.Type() == core.TYPE_STREAM {
		s.attachPort(p.Stream(), key)
	}
}

func (p *storePipe) pull() interface{} {
	if p.index >= len(p.items) {
		return core.PHMultiple
	}
	index := p.index
	p.index++
	return p.items[index]
}

func (s store) pull(p *core.Port, key string) interface{} {
	if p.Primitive() {
		return s[key].pull()
	} else if p.Type() == core.TYPE_MAP {
		obj := make(map[string]interface{})
		for _, sub := range p.MapEntries() {
			obj[sub] = s.pull(p.Map(sub), key+"."+sub)
		}
		newObj := false
		var marker interface{} = nil
		for sub := range obj {
			if obj[sub] != core.PHMultiple && !core.IsMarker(obj[sub]) {
				if marker != nil {
					panic("markers not matching 1")
				}
				newObj = true
				continue
			}
			if obj[sub] == core.PHMultiple {
				obj[sub] = core.PHSingle
				continue
			}
			if core.IsMarker(obj[sub]) {
				if marker != nil && marker != obj[sub] {
					panic("markers not matching 2")
				}
				marker = obj[sub]
				continue
			} else if marker != nil {
				panic("markers not matching 3")
			}
		}
		if marker != nil {
			return marker
		}
		if newObj {
			return obj
		} else {
			return core.PHMultiple
		}
	} else if p.Type() == core.TYPE_STREAM {
		bos := s.pull(p.Stream(), key)
		if bos == core.PHMultiple || !p.OwnBOS(bos) {
			return bos
		}
		obj := []interface{}{}
		for {
			el := s.pull(p.Stream(), key)
			if el == core.PHMultiple {
				obj = append(obj, core.PHMultiple)
				break
			}
			if p.OwnEOS(el) {
				break
			}
			obj = append(obj, el)
		}
		return obj
	}

	return core.PHMultiple
}

func (s store) resetIndexes() {
	for pipe := range s {
		s[pipe].index = 0
	}
}

var metaStoreCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type:    "generic",
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
					Type: "stream",
					Stream: &core.TypeDef{
						Type:    "generic",
						Generic: "examineType",
					},
				},
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		//out := op.Main().Out()
		querySrv := op.Service("query")
		queryIn := querySrv.In()
		queryOut := querySrv.Out()

		store := make(store)
		store.attachPort(in, "in")

		for !op.CheckStop() {
			queryIn.Pull()
			store.resetIndexes()
			obj := []interface{}{}
			for {
				el := store.pull(in, "in")
				if el == core.PHMultiple {
					break
				}
				obj = append(obj, el)
			}
			queryOut.Push(obj)
		}
	},
}
