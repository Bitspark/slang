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
	o1, _ := core.NewOperator("o1", nil, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: defIn, Out: defOut}}, nil)

	o1.Main().In().Connect(o1.Main().Out())

	o1.Main().Out().Bufferize()
	o1.Main().In().Push(1.0)

	a.PortPushesAll(parseJSON(`[1]`).([]interface{}), o1.Main().Out())
}

func TestNetwork_EmptyOperators(t *testing.T) {
	a := assertions.New(t)
	defIn := api.ParsePortDef(`{"type":"number"}`)
	defOut := api.ParsePortDef(`{"type":"number"}`)
	o1, _ := core.NewOperator("o1", nil, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: defIn, Out: defOut}}, nil)
	o2, _ := core.NewOperator("o2", nil, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: defIn, Out: defOut}}, nil)
	o2.SetParent(o1)
	o3, _ := core.NewOperator("o3", nil, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: defIn, Out: defOut}}, nil)
	o3.SetParent(o2)
	o4, _ := core.NewOperator("o4", nil, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: defIn, Out: defOut}}, nil)
	o4.SetParent(o2)

	o3.Main().In().Connect(o3.Main().Out())
	o4.Main().In().Connect(o4.Main().Out())
	o2.Main().In().Connect(o3.Main().In())
	o3.Main().Out().Connect(o4.Main().In())
	o4.Main().Out().Connect(o2.Main().Out())
	o1.Main().In().Connect(o2.Main().In())
	o2.Main().Out().Connect(o1.Main().Out())

	if o1.Main().In().Connected(o1.Main().Out()) {
		t.Error("should not be connected")
	}

	if !o1.Main().In().Connected(o2.Main().In()) {
		t.Error("should be connected")
	}

	o3.Compile()
	o4.Compile()
	o2.Compile()

	if !o1.Main().In().Connected(o1.Main().Out()) {
		t.Error("should be connected")
	}

	if o1.Main().In().Connected(o2.Main().In()) {
		t.Error("should not be connected")
	}

	o1.Main().Out().Bufferize()
	o1.Main().In().Push(1.0)

	a.PortPushesAll(parseJSON(`[1]`).([]interface{}), o1.Main().Out())
}

func TestNetwork_DoubleSum(t *testing.T) {
	a := assertions.New(t)
	defStrStr := api.ParsePortDef(`{"type":"stream","stream":{"type":"stream","stream":{"type":"number"}}}`)
	defStr := api.ParsePortDef(`{"type":"stream","stream":{"type":"number"}}`)
	def := api.ParsePortDef(`{"type":"number"}`)

	double := func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for {
			i := in.Pull()
			if n, ok := i.(float64); ok {
				out.Push(2 * n)
			} else {
				out.Push(i)
			}
		}
	}

	sum := func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for {
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

	o1, _ := core.NewOperator("O1", nil, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: defStrStr, Out: defStr}}, nil)
	o2, _ := core.NewOperator("O2", nil, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: defStr, Out: def}}, nil)
	o2.SetParent(o1)
	o3, _ := core.NewOperator("O3", double, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: def, Out: def}}, nil)
	o3.SetParent(o2)
	o4, _ := core.NewOperator("O4", sum, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: defStr, Out: def}}, nil)
	o4.SetParent(o2)

	err := o2.Main().In().Stream().Connect(o3.Main().In())
	a.NoError(err)
	err = o3.Main().Out().Connect(o4.Main().In().Stream())
	a.NoError(err)
	err = o4.Main().Out().Connect(o2.Main().Out())
	a.NoError(err)

	if !o2.Main().In().Stream().Connected(o3.Main().In()) {
		t.Error("should be connected")
	}

	if !o3.Main().Out().Connected(o4.Main().In().Stream()) {
		t.Error("should be connected")
	}

	if !o4.Main().Out().Connected(o2.Main().Out()) {
		t.Error("should be connected")
	}

	if o3.BasePort() != o2.Main().In() {
		t.Error("wrong base port")
	}

	if !o2.Main().In().Connected(o4.Main().In()) {
		t.Error("should be connected via base port")
	}

	//

	err = o1.Main().In().Stream().Stream().Connect(o2.Main().In().Stream())
	a.NoError(err)
	err = o2.Main().Out().Connect(o1.Main().Out().Stream())
	a.NoError(err)

	if !o1.Main().In().Stream().Stream().Connected(o2.Main().In().Stream()) {
		t.Error("should be connected")
	}

	if !o1.Main().In().Stream().Connected(o2.Main().In()) {
		t.Error("should be connected")
	}

	if !o2.Main().Out().Connected(o1.Main().Out().Stream()) {
		t.Error("should be connected")
	}

	if o2.BasePort() != o1.Main().In() {
		t.Error("wrong base port")
	}

	if !o1.Main().In().Connected(o1.Main().Out()) {
		t.Error("should be connected via base port")
	}

	//

	o2.Compile()

	if !o1.Main().In().Connected(o1.Main().Out()) {
		t.Error("should be connected")
	}

	if !o1.Main().In().Stream().Connected(o4.Main().In()) {
		t.Error("should be connected")
	}

	if !o1.Main().In().Stream().Stream().Connected(o3.Main().In()) {
		t.Error("should be connected")
	}

	if !o3.Main().Out().Connected(o4.Main().In().Stream()) {
		t.Error("should be connected")
	}

	//

	o1.Main().Out().Bufferize()

	go o3.Start()
	go o4.Start()

	o1.Main().In().Push(parseJSON(`[[1,2,3],[4,5]]`))
	o1.Main().In().Push(parseJSON(`[[],[2]]`))
	o1.Main().In().Push(parseJSON(`[]`))
	a.PortPushesAll(parseJSON(`[[12,18],[0,4],[]]`).([]interface{}), o1.Main().Out())
}

func TestNetwork_NumgenSum(t *testing.T) {
	a := assertions.New(t)
	defStrStrStr := api.ParsePortDef(`{"type":"stream","stream":{"type":"stream","stream":{"type":"stream","stream":{"type":"number"}}}}`)
	defStrStr := api.ParsePortDef(`{"type":"stream","stream":{"type":"stream","stream":{"type":"number"}}}`)
	defStr := api.ParsePortDef(`{"type":"stream","stream":{"type":"number"}}`)
	def := api.ParsePortDef(`{"type":"number"}`)

	numgen := func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for {
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

	sum := func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for {
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

	o1, _ := core.NewOperator("O1", nil, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: defStr, Out: defStrStr}}, nil)
	o2, _ := core.NewOperator("O2", numgen, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: def, Out: defStr}}, nil)
	o2.SetParent(o1)
	o3, _ := core.NewOperator("O3", numgen, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: def, Out: defStr}}, nil)
	o3.SetParent(o1)
	o4, _ := core.NewOperator("O4", nil, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: defStrStrStr, Out: defStrStr}}, nil)
	o4.SetParent(o1)
	o5, _ := core.NewOperator("O5", sum, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: defStr, Out: def}}, nil)
	o5.SetParent(o4)

	o4.Main().In().Stream().Stream().Stream().Connect(o5.Main().In().Stream())
	o5.Main().Out().Connect(o4.Main().Out().Stream().Stream())

	if !o4.Main().In().Stream().Stream().Stream().Connected(o5.Main().In().Stream()) {
		t.Error("should be connected")
	}

	if !o4.Main().In().Stream().Stream().Connected(o5.Main().In()) {
		t.Error("should be connected")
	}

	if !o5.Main().Out().Connected(o4.Main().Out().Stream().Stream()) {
		t.Error("should be connected")
	}

	if !o4.Main().In().Stream().Connected(o4.Main().Out().Stream()) {
		t.Error("should be connected via base port")
	}

	if !o4.Main().In().Connected(o4.Main().Out()) {
		t.Error("should be connected via base port")
	}

	//

	o1.Main().In().Stream().Connect(o2.Main().In())
	o2.Main().Out().Stream().Connect(o3.Main().In())
	o3.Main().Out().Stream().Connect(o4.Main().In().Stream().Stream().Stream())
	o4.Main().Out().Stream().Stream().Connect(o1.Main().Out().Stream().Stream())

	if !o1.Main().In().Stream().Connected(o2.Main().In()) {
		t.Error("should be connected")
	}

	if !o2.Main().Out().Stream().Connected(o3.Main().In()) {
		t.Error("should be connected")
	}

	if !o3.Main().Out().Stream().Connected(o4.Main().In().Stream().Stream().Stream()) {
		t.Error("should be connected")
	}

	if !o3.Main().Out().Connected(o4.Main().In().Stream().Stream()) {
		t.Error("should be connected")
	}

	if !o4.Main().Out().Stream().Stream().Connected(o1.Main().Out().Stream().Stream()) {
		t.Error("should be connected")
	}

	if !o4.Main().Out().Stream().Connected(o1.Main().Out().Stream()) {
		t.Error("should be connected")
	}

	if !o4.Main().Out().Connected(o1.Main().Out()) {
		t.Error("should be connected")
	}

	if o2.BasePort() != o1.Main().In() {
		t.Error("wrong base port")
	}

	if o3.BasePort() != o2.Main().Out() {
		t.Error("wrong base port")
	}

	if !o1.Main().In().Connected(o4.Main().In()) {
		t.Error("should be connected via base port")
	}

	if !o2.Main().Out().Connected(o4.Main().In().Stream()) {
		t.Error("should be connected via base port")
	}

	//

	o4.Compile()

	if !o1.Main().In().Connected(o1.Main().Out()) {
		t.Error("should be connected after merge")
	}

	if !o2.Main().Out().Connected(o1.Main().Out().Stream()) {
		t.Error("should be connected after merge")
	}

	if !o3.Main().Out().Stream().Connected(o5.Main().In().Stream()) {
		t.Error("should be connected after merge")
	}

	if !o3.Main().Out().Connected(o5.Main().In()) {
		t.Error("should be connected after merge")
	}

	//

	o1.Main().Out().Bufferize()

	go o2.Start()
	go o3.Start()
	go o5.Start()

	o1.Main().In().Push(parseJSON(`[1,2,3]`))
	o1.Main().In().Push(parseJSON(`[]`))
	o1.Main().In().Push(parseJSON(`[4]`))
	a.PortPushesAll(parseJSON(`[[[1],[1,3],[1,3,6]],[],[[1,3,6,10]]]`).([]interface{}), o1.Main().Out())
}

func TestNetwork_Maps_Simple(t *testing.T) {
	a := assertions.New(t)
	defIn := api.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"},"b":{"type":"number"}}}`)
	defOut := defIn

	defMap1In := api.ParsePortDef(`{"type":"number"}`)
	defMap1Out := api.ParsePortDef(`{"type":"map","map":{"a":{"type":"number"},"b":{"type":"number"}}}`)

	defMap2In := defMap1Out
	defMap2Out := defMap1In

	evalMap1 := func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for {
			i := in.Pull()
			if i, ok := i.(float64); ok {
				out.Map("a").Push(2 * i)
				out.Map("b").Push(3 * i)
			} else {
				out.Push(i)
			}
		}
	}

	evalMap2 := func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for {
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

	o, _ := core.NewOperator("", nil, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: defIn, Out: defOut}}, nil)
	oMap1, _ := core.NewOperator("Map1", evalMap1, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: defMap1In, Out: defMap1Out}}, nil)
	oMap1.SetParent(o)
	oMap2, _ := core.NewOperator("Map2", evalMap2, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: defMap2In, Out: defMap2Out}}, nil)
	oMap2.SetParent(o)

	o.Main().In().Map("a").Connect(oMap2.Main().In().Map("a"))
	o.Main().In().Map("b").Connect(oMap1.Main().In())
	oMap1.Main().Out().Map("a").Connect(oMap2.Main().In().Map("b"))
	oMap1.Main().Out().Map("b").Connect(o.Main().Out().Map("b"))
	oMap2.Main().Out().Connect(o.Main().Out().Map("a"))

	o.Main().Out().Bufferize()

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
		o.Main().In().Push(parseJSON(d))
	}

	a.PortPushesAll(parseJSON(results).([]interface{}), o.Main().Out())

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

	sumEval := func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for {
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

	filterEval := func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for {
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

	addEval := func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for {
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

	o, _ := core.NewOperator("Global", nil, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: defStrMapStr, Out: defStrMap}}, nil)
	sum, _ := core.NewOperator("Sum", sumEval, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: defSumIn, Out: defSumOut}}, nil)
	sum.SetParent(o)
	add, _ := core.NewOperator("Add", addEval, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: defAddIn, Out: defAddOut}}, nil)
	add.SetParent(o)
	filter1, _ := core.NewOperator("Filter1", filterEval, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: defFilterIn, Out: defFilterOut}}, nil)
	filter1.SetParent(o)
	filter2, _ := core.NewOperator("Filter2", filterEval, nil, map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: defFilterIn, Out: defFilterOut}}, nil)
	filter2.SetParent(o)

	o.Main().In().Stream().Map("N").Connect(sum.Main().In())
	o.Main().In().Stream().Map("n").Connect(add.Main().In().Map("b"))
	o.Main().In().Stream().Map("b").Connect(filter1.Main().In().Map("b"))
	o.Main().In().Stream().Map("b").Connect(filter2.Main().In().Map("b"))
	o.Main().In().Stream().Map("s").Connect(filter2.Main().In().Map("o"))
	sum.Main().Out().Connect(add.Main().In().Map("a"))
	add.Main().Out().Connect(filter1.Main().In().Map("o"))
	filter1.Main().Out().Connect(o.Main().Out().Stream().Map("sum"))
	filter2.Main().Out().Connect(o.Main().Out().Stream().Map("s"))

	o.Main().Out().Bufferize()

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
		o.Main().In().Push(parseJSON(d))
	}

	a.PortPushesAll(parseJSON(results).([]interface{}), o.Main().Out())
}
