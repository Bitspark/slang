package tests

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"testing"
	"github.com/Bitspark/slang/pkg/api"
)

func TestNetwork_EmptyOperator(t *testing.T) {
	a := assertions.New(t)
	defIn := api.ParsePortDef(`{"type":"number"}`)
	defOut := api.ParsePortDef(`{"type":"number"}`)
	o1, _ := core.NewOperator("o1", nil, defIn, defOut, nil)

	o1.In().Connect(o1.Out())

	o1.Out().Bufferize()
	o1.In().Push(1.0)

	a.PortPushes(parseJSON(`[1]`).([]interface{}), o1.Out())
}

func TestNetwork_EmptyOperators(t *testing.T) {
	a := assertions.New(t)
	defIn := api.ParsePortDef(`{"type":"number"}`)
	defOut := api.ParsePortDef(`{"type":"number"}`)
	o1, _ := core.NewOperator("o1", nil, defIn, defOut, nil)
	o2, _ := core.NewOperator("o2", nil, defIn, defOut, nil)
	o2.SetParent(o1)
	o3, _ := core.NewOperator("o3", nil, defIn, defOut, nil)
	o3.SetParent(o2)
	o4, _ := core.NewOperator("o4", nil, defIn, defOut, nil)
	o4.SetParent(o2)

	o3.In().Connect(o3.Out())
	o4.In().Connect(o4.Out())
	o2.In().Connect(o3.In())
	o3.Out().Connect(o4.In())
	o4.Out().Connect(o2.Out())
	o1.In().Connect(o2.In())
	o2.Out().Connect(o1.Out())

	if o1.In().Connected(o1.Out()) {
		t.Error("should not be connected")
	}

	if !o1.In().Connected(o2.In()) {
		t.Error("should be connected")
	}

	o3.Compile()
	o4.Compile()
	o2.Compile()

	if !o1.In().Connected(o1.Out()) {
		t.Error("should be connected")
	}

	if o1.In().Connected(o2.In()) {
		t.Error("should not be connected")
	}

	o1.Out().Bufferize()
	o1.In().Push(1.0)

	a.PortPushes(parseJSON(`[1]`).([]interface{}), o1.Out())
}

func TestNetwork_DoubleSum(t *testing.T) {
	a := assertions.New(t)
	defStrStr := api.ParsePortDef(`{"type":"stream","stream":{"type":"stream","stream":{"type":"number"}}}`)
	defStr := api.ParsePortDef(`{"type":"stream","stream":{"type":"number"}}`)
	def := api.ParsePortDef(`{"type":"number"}`)

	double := func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
		for true {
			i := in.Pull()
			if n, ok := i.(float64); ok {
				out.Push(2 * n)
			} else {
				out.Push(i)
			}
		}
	}

	sum := func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
		for true {
			i := in.Pull()
			if ns, ok := i.([]interface{}); ok {
				sum := 0.0
				for _, n := range ns {
					sum += n.(float64)
				}
				out.Push(sum)
			} else {
				out.Push(i)
			}
		}
	}

	o1, _ := core.NewOperator("O1", nil, defStrStr, defStr, nil)
	o2, _ := core.NewOperator("O2", nil, defStr, def, nil)
	o2.SetParent(o1)
	o3, _ := core.NewOperator("O3", double, def, def, nil)
	o3.SetParent(o2)
	o4, _ := core.NewOperator("O4", sum, defStr, def, nil)
	o4.SetParent(o2)

	err := o2.In().Stream().Connect(o3.In())
	a.NoError(err)
	err = o3.Out().Connect(o4.In().Stream())
	a.NoError(err)
	err = o4.Out().Connect(o2.Out())
	a.NoError(err)

	if !o2.In().Stream().Connected(o3.In()) {
		t.Error("should be connected")
	}

	if !o3.Out().Connected(o4.In().Stream()) {
		t.Error("should be connected")
	}

	if !o4.Out().Connected(o2.Out()) {
		t.Error("should be connected")
	}

	if o3.BasePort() != o2.In() {
		t.Error("wrong base port")
	}

	if !o2.In().Connected(o4.In()) {
		t.Error("should be connected via base port")
	}

	//

	err = o1.In().Stream().Stream().Connect(o2.In().Stream())
	a.NoError(err)
	err = o2.Out().Connect(o1.Out().Stream())
	a.NoError(err)

	if !o1.In().Stream().Stream().Connected(o2.In().Stream()) {
		t.Error("should be connected")
	}

	if !o1.In().Stream().Connected(o2.In()) {
		t.Error("should be connected")
	}

	if !o2.Out().Connected(o1.Out().Stream()) {
		t.Error("should be connected")
	}

	if o2.BasePort() != o1.In() {
		t.Error("wrong base port")
	}

	if !o1.In().Connected(o1.Out()) {
		t.Error("should be connected via base port")
	}

	//

	o2.Compile()

	if !o1.In().Connected(o1.Out()) {
		t.Error("should be connected")
	}

	if !o1.In().Stream().Connected(o4.In()) {
		t.Error("should be connected")
	}

	if !o1.In().Stream().Stream().Connected(o3.In()) {
		t.Error("should be connected")
	}

	if !o3.Out().Connected(o4.In().Stream()) {
		t.Error("should be connected")
	}

	//

	o1.Out().Bufferize()

	go o3.Start()
	go o4.Start()

	o1.In().Push(parseJSON(`[[1,2,3],[4,5]]`))
	o1.In().Push(parseJSON(`[[],[2]]`))
	o1.In().Push(parseJSON(`[]`))
	a.PortPushes(parseJSON(`[[12,18],[0,4],[]]`).([]interface{}), o1.Out())
}

func TestNetwork_NumgenSum(t *testing.T) {
	a := assertions.New(t)
	defStrStrStr := api.ParsePortDef(`{"type":"stream","stream":{"type":"stream","stream":{"type":"stream","stream":{"type":"number"}}}}`)
	defStrStr := api.ParsePortDef(`{"type":"stream","stream":{"type":"stream","stream":{"type":"number"}}}`)
	defStr := api.ParsePortDef(`{"type":"stream","stream":{"type":"number"}}`)
	def := api.ParsePortDef(`{"type":"number"}`)

	numgen := func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
		for true {
			i := in.Pull()
			if n, ok := i.(float64); ok {
				ns := []interface{}{}
				for i := 1; i <= int(n); i++ {
					ns = append(ns, float64(i))
				}
				out.Push(ns)
			} else {
				out.Push(i)
			}
		}
	}

	sum := func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
		for true {
			i := in.Pull()
			if ns, ok := i.([]interface{}); ok {
				sum := 0.0
				for _, n := range ns {
					sum += n.(float64)
				}
				out.Push(sum)
			} else {
				out.Push(i)
			}
		}
	}

	o1, _ := core.NewOperator("O1", nil, defStr, defStrStr, nil)
	o2, _ := core.NewOperator("O2", numgen, def, defStr, nil)
	o2.SetParent(o1)
	o3, _ := core.NewOperator("O3", numgen, def, defStr, nil)
	o3.SetParent(o1)
	o4, _ := core.NewOperator("O4", nil, defStrStrStr, defStrStr, nil)
	o4.SetParent(o1)
	o5, _ := core.NewOperator("O5", sum, defStr, def, nil)
	o5.SetParent(o4)

	o4.In().Stream().Stream().Stream().Connect(o5.In().Stream())
	o5.Out().Connect(o4.Out().Stream().Stream())

	if !o4.In().Stream().Stream().Stream().Connected(o5.In().Stream()) {
		t.Error("should be connected")
	}

	if !o4.In().Stream().Stream().Connected(o5.In()) {
		t.Error("should be connected")
	}

	if !o5.Out().Connected(o4.Out().Stream().Stream()) {
		t.Error("should be connected")
	}

	if !o4.In().Stream().Connected(o4.Out().Stream()) {
		t.Error("should be connected via base port")
	}

	if !o4.In().Connected(o4.Out()) {
		t.Error("should be connected via base port")
	}

	//

	o1.In().Stream().Connect(o2.In())
	o2.Out().Stream().Connect(o3.In())
	o3.Out().Stream().Connect(o4.In().Stream().Stream().Stream())
	o4.Out().Stream().Stream().Connect(o1.Out().Stream().Stream())

	if !o1.In().Stream().Connected(o2.In()) {
		t.Error("should be connected")
	}

	if !o2.Out().Stream().Connected(o3.In()) {
		t.Error("should be connected")
	}

	if !o3.Out().Stream().Connected(o4.In().Stream().Stream().Stream()) {
		t.Error("should be connected")
	}

	if !o3.Out().Connected(o4.In().Stream().Stream()) {
		t.Error("should be connected")
	}

	if !o4.Out().Stream().Stream().Connected(o1.Out().Stream().Stream()) {
		t.Error("should be connected")
	}

	if !o4.Out().Stream().Connected(o1.Out().Stream()) {
		t.Error("should be connected")
	}

	if !o4.Out().Connected(o1.Out()) {
		t.Error("should be connected")
	}

	if o2.BasePort() != o1.In() {
		t.Error("wrong base port")
	}

	if o3.BasePort() != o2.Out() {
		t.Error("wrong base port")
	}

	if !o1.In().Connected(o4.In()) {
		t.Error("should be connected via base port")
	}

	if !o2.Out().Connected(o4.In().Stream()) {
		t.Error("should be connected via base port")
	}

	//

	o4.Compile()

	if !o1.In().Connected(o1.Out()) {
		t.Error("should be connected after merge")
	}

	if !o2.Out().Connected(o1.Out().Stream()) {
		t.Error("should be connected after merge")
	}

	if !o3.Out().Stream().Connected(o5.In().Stream()) {
		t.Error("should be connected after merge")
	}

	if !o3.Out().Connected(o5.In()) {
		t.Error("should be connected after merge")
	}

	//

	o1.Out().Bufferize()

	go o2.Start()
	go o3.Start()
	go o5.Start()

	o1.In().Push(parseJSON(`[1,2,3]`))
	o1.In().Push(parseJSON(`[]`))
	o1.In().Push(parseJSON(`[4]`))
	a.PortPushes(parseJSON(`[[[1],[1,3],[1,3,6]],[],[[1,3,6,10]]]`).([]interface{}), o1.Out())
}

func TestNetwork_Maps_Simple(t *testing.T) {
	a := assertions.New(t)
	defIn := api.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"},"b":{"type":"number"}}}`)
	defOut := defIn

	defMap1In := api.ParsePortDef(`{"type":"number"}`)
	defMap1Out := api.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"},"b":{"type":"number"}}}`)

	defMap2In := defMap1Out
	defMap2Out := defMap1In

	evalMap1 := func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
		for true {
			i := in.Pull()
			if i, ok := i.(float64); ok {
				out.Map("a").Push(2 * i)
				out.Map("b").Push(3 * i)
			} else {
				out.Push(i)
			}
		}
	}

	evalMap2 := func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
		for true {
			i := in.Pull()
			if m, ok := i.(map[string]interface{}); ok {
				a := m["a"].(float64)
				b := m["b"].(float64)
				out.Push(a * b)
			} else {
				out.Push(i)
			}
		}
	}

	o, _ := core.NewOperator("", nil, defIn, defOut, nil)
	oMap1, _ := core.NewOperator("Map1", evalMap1, defMap1In, defMap1Out, nil)
	oMap1.SetParent(o)
	oMap2, _ := core.NewOperator("Map2", evalMap2, defMap2In, defMap2Out, nil)
	oMap2.SetParent(o)

	o.In().Map("a").Connect(oMap2.In().Map("a"))
	o.In().Map("b").Connect(oMap1.In())
	oMap1.Out().Map("a").Connect(oMap2.In().Map("b"))
	oMap1.Out().Map("b").Connect(o.Out().Map("b"))
	oMap2.Out().Connect(o.Out().Map("a"))

	o.Out().Bufferize()

	go oMap1.Start()
	go oMap2.Start()

	dataIn := []string{
		`{"a":1,"b":1}`,
		`{"a":1,"b":0}`,
		`{"a":0,"b":1}`,
		`{"a":2,"b":3}`,
	}
	results := `[{"a":2,"b":3},{"a":0,"b":0},{"a":0,"b":3},{"a":12,"b":9}]`

	for _, d := range dataIn {
		o.In().Push(parseJSON(d))
	}

	a.PortPushes(parseJSON(results).([]interface{}), o.Out())

}

func TestNetwork_Maps_Complex(t *testing.T) {
	a := assertions.New(t)
	defStrMapStr := api.ParsePortDef(`{"type":"stream","stream":{"type":"map","map":{
		"N":{"type":"stream","stream":{"type":"number"}},
		"n":{"type":"number"},
		"s":{"type":"string"},
		"b":{"type":"boolean"}}}}`)
	defStrMap := api.ParsePortDef(`{"type":"stream","stream":{"type":"map","map":{
		"sum":{"type":"number"},
		"s":{"type":"string"}}}}`)
	defFilterIn := api.ParsePortDef(`{"type":"map","map":{
		"o":{"type":"primitive"},
		"b":{"type":"boolean"}}}`)
	defFilterOut := api.ParsePortDef(`{"type":"primitive"}`)
	defAddIn := api.ParsePortDef(`{"type":"map","map":{
		"a":{"type":"number"},
		"b":{"type":"number"}}}`)
	defAddOut := api.ParsePortDef(`{"type":"number"}`)
	defSumIn := api.ParsePortDef(`{"type":"stream","stream":{"type":"number"}}`)
	defSumOut := api.ParsePortDef(`{"type":"number"}`)

	sumEval := func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
		for true {
			i := in.Pull()
			if ns, ok := i.([]interface{}); ok {
				sum := 0.0
				for _, n := range ns {
					sum += n.(float64)
				}
				out.Push(sum)
			} else {
				out.Push(i)
			}
		}
	}

	filterEval := func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
		for true {
			i := in.Pull()
			if m, ok := i.(map[string]interface{}); ok {
				if m["b"].(bool) {
					out.Push(m["o"])
				}
			} else {
				out.Push(i)
			}
		}
	}

	addEval := func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
		for true {
			i := in.Pull()
			if m, ok := i.(map[string]interface{}); ok {
				a := m["a"].(float64)
				b := m["b"].(float64)
				out.Push(a + b)
			} else {
				out.Push(i)
			}
		}
	}

	o, _ := core.NewOperator("Global", nil, defStrMapStr, defStrMap, nil)
	sum, _ := core.NewOperator("Sum", sumEval, defSumIn, defSumOut, nil)
	sum.SetParent(o)
	add, _ := core.NewOperator("Add", addEval, defAddIn, defAddOut, nil)
	add.SetParent(o)
	filter1, _ := core.NewOperator("Filter1", filterEval, defFilterIn, defFilterOut, nil)
	filter1.SetParent(o)
	filter2, _ := core.NewOperator("Filter2", filterEval, defFilterIn, defFilterOut, nil)
	filter2.SetParent(o)

	o.In().Stream().Map("N").Connect(sum.In())
	o.In().Stream().Map("n").Connect(add.In().Map("b"))
	o.In().Stream().Map("b").Connect(filter1.In().Map("b"))
	o.In().Stream().Map("b").Connect(filter2.In().Map("b"))
	o.In().Stream().Map("s").Connect(filter2.In().Map("o"))
	sum.Out().Connect(add.In().Map("a"))
	add.Out().Connect(filter1.In().Map("o"))
	filter1.Out().Connect(o.Out().Stream().Map("sum"))
	filter2.Out().Connect(o.Out().Stream().Map("s"))

	o.Out().Bufferize()

	go sum.Start()
	go add.Start()
	go filter1.Start()
	go filter2.Start()

	dataIn := []string{
		`[{"N":[1,2,4],"n":2,"s":"must pass","b":true},{"N":[4,5],"n":-6,"s":"","b":true}]`,
		`[{"N":[10,20,40],"n":20,"s":"may not pass","b":false}]`,
		`[]`,
		`[{"N":[],"n":1,"s":"must also pass","b":true}]`,
	}
	results := `[[{"sum":9,"s":"must pass"},{"sum":3,"s":""}],[],[],[{"sum":1,"s":"must also pass"}]]`

	for _, d := range dataIn {
		o.In().Push(parseJSON(d))
	}

	a.PortPushes(parseJSON(results).([]interface{}), o.Out())
}
