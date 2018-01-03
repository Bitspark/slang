package slang

import (
	"slang/op"
	"testing"
)

func TestNetwork_EmptyOperator(t *testing.T) {
	defIn := helperJson2PortDef(`{"type":"number"}`)
	defOut := helperJson2PortDef(`{"type":"number"}`)
	o1, _ := op.MakeOperator("o1", nil, defIn, defOut, nil)

	o1.InPort().Connect(o1.OutPort())

	o1.OutPort().Bufferize()
	o1.InPort().Push(1.0)

	assertPortItems(t, helperJson2I(`[1]`).([]interface{}), o1.OutPort())
}

func TestNetwork_EmptyOperators(t *testing.T) {
	defIn := helperJson2PortDef(`{"type":"number"}`)
	defOut := helperJson2PortDef(`{"type":"number"}`)
	o1, _ := op.MakeOperator("o1", nil, defIn, defOut, nil)
	o2, _ := op.MakeOperator("o2", nil, defIn, defOut, o1)
	o3, _ := op.MakeOperator("o3", nil, defIn, defOut, o2)
	o4, _ := op.MakeOperator("o4", nil, defIn, defOut, o2)

	o3.InPort().Connect(o3.OutPort())
	o4.InPort().Connect(o4.OutPort())
	o2.InPort().Connect(o3.InPort())
	o3.OutPort().Connect(o4.InPort())
	o4.OutPort().Connect(o2.OutPort())
	o1.InPort().Connect(o2.InPort())
	o2.OutPort().Connect(o1.OutPort())

	if o1.InPort().Connected(o1.OutPort()) {
		t.Error("should not be connected")
	}

	if !o1.InPort().Connected(o2.InPort()) {
		t.Error("should be connected")
	}

	o3.Compile()
	o4.Compile()
	o2.Compile()

	if !o1.InPort().Connected(o1.OutPort()) {
		t.Error("should be connected")
	}

	if o1.InPort().Connected(o2.InPort()) {
		t.Error("should not be connected")
	}

	o1.OutPort().Bufferize()
	o1.InPort().Push(1.0)

	assertPortItems(t, helperJson2I(`[1]`).([]interface{}), o1.OutPort())
}

func TestNetwork_DoubleSum(t *testing.T) {
	defStrStr := helperJson2PortDef(`{"type":"stream","stream":{"type":"stream","stream":{"type":"number"}}}`)
	defStr := helperJson2PortDef(`{"type":"stream","stream":{"type":"number"}}`)
	def := helperJson2PortDef(`{"type":"number"}`)

	double := func(in, out *op.Port, store interface{}) {
		for true {
			i := in.Pull()
			if n, ok := i.(float64); ok {
				out.Push(2 * n)
			} else {
				out.Push(i)
			}
		}
	}

	sum := func(in, out *op.Port, store interface{}) {
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

	o1, _ := op.MakeOperator("O1", nil, defStrStr, defStr, nil)
	o2, _ := op.MakeOperator("O2", nil, defStr, def, o1)
	o3, _ := op.MakeOperator("O3", double, def, def, o2)
	o4, _ := op.MakeOperator("O4", sum, defStr, def, o2)

	err := o2.InPort().Stream().Connect(o3.InPort())
	assertNoError(t, err)
	err = o3.OutPort().Connect(o4.InPort().Stream())
	assertNoError(t, err)
	err = o4.OutPort().Connect(o2.OutPort())
	assertNoError(t, err)

	if !o2.InPort().Stream().Connected(o3.InPort()) {
		t.Error("should be connected")
	}

	if !o3.OutPort().Connected(o4.InPort().Stream()) {
		t.Error("should be connected")
	}

	if !o4.OutPort().Connected(o2.OutPort()) {
		t.Error("should be connected")
	}

	if o3.BasePort() != o2.InPort() {
		t.Error("wrong base port")
	}

	if !o2.InPort().Connected(o4.InPort()) {
		t.Error("should be connected via base port")
	}

	//

	err = o1.InPort().Stream().Stream().Connect(o2.InPort().Stream())
	assertNoError(t, err)
	err = o2.OutPort().Connect(o1.OutPort().Stream())
	assertNoError(t, err)

	if !o1.InPort().Stream().Stream().Connected(o2.InPort().Stream()) {
		t.Error("should be connected")
	}

	if !o1.InPort().Stream().Connected(o2.InPort()) {
		t.Error("should be connected")
	}

	if !o2.OutPort().Connected(o1.OutPort().Stream()) {
		t.Error("should be connected")
	}

	if o2.BasePort() != o1.InPort() {
		t.Error("wrong base port")
	}

	if !o1.InPort().Connected(o1.OutPort()) {
		t.Error("should be connected via base port")
	}

	//

	o2.Compile()

	if !o1.InPort().Connected(o1.OutPort()) {
		t.Error("should be connected")
	}

	if !o1.InPort().Stream().Connected(o4.InPort()) {
		t.Error("should be connected")
	}

	if !o1.InPort().Stream().Stream().Connected(o3.InPort()) {
		t.Error("should be connected")
	}

	if !o3.OutPort().Connected(o4.InPort().Stream()) {
		t.Error("should be connected")
	}

	//

	o1.OutPort().Stream().Bufferize()

	go o3.Start()
	go o4.Start()

	o1.InPort().Push(helperJson2I(`[[1,2,3],[4,5]]`))
	o1.InPort().Push(helperJson2I(`[[],[2]]`))
	o1.InPort().Push(helperJson2I(`[]`))
	assertPortItems(t, helperJson2I(`[[12,18],[0,4],[]]`).([]interface{}), o1.OutPort())
}

func TestNetwork_NumgenSum(t *testing.T) {
	defStrStrStr := helperJson2PortDef(`{"type":"stream","stream":{"type":"stream","stream":{"type":"stream","stream":{"type":"number"}}}}`)
	defStrStr := helperJson2PortDef(`{"type":"stream","stream":{"type":"stream","stream":{"type":"number"}}}`)
	defStr := helperJson2PortDef(`{"type":"stream","stream":{"type":"number"}}`)
	def := helperJson2PortDef(`{"type":"number"}`)

	numgen := func(in, out *op.Port, store interface{}) {
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

	sum := func(in, out *op.Port, store interface{}) {
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

	o1, _ := op.MakeOperator("O1", nil, defStr, defStrStr, nil)
	o2, _ := op.MakeOperator("O2", numgen, def, defStr, o1)
	o3, _ := op.MakeOperator("O3", numgen, def, defStr, o1)
	o4, _ := op.MakeOperator("O4", nil, defStrStrStr, defStrStr, o1)
	o5, _ := op.MakeOperator("O5", sum, defStr, def, o4)

	o4.InPort().Stream().Stream().Stream().Connect(o5.InPort().Stream())
	o5.OutPort().Connect(o4.OutPort().Stream().Stream())

	if !o4.InPort().Stream().Stream().Stream().Connected(o5.InPort().Stream()) {
		t.Error("should be connected")
	}

	if !o4.InPort().Stream().Stream().Connected(o5.InPort()) {
		t.Error("should be connected")
	}

	if !o5.OutPort().Connected(o4.OutPort().Stream().Stream()) {
		t.Error("should be connected")
	}

	if !o4.InPort().Stream().Connected(o4.OutPort().Stream()) {
		t.Error("should be connected via base port")
	}

	if !o4.InPort().Connected(o4.OutPort()) {
		t.Error("should be connected via base port")
	}

	//

	o1.InPort().Stream().Connect(o2.InPort())
	o2.OutPort().Stream().Connect(o3.InPort())
	o3.OutPort().Stream().Connect(o4.InPort().Stream().Stream().Stream())
	o4.OutPort().Stream().Stream().Connect(o1.OutPort().Stream().Stream())

	if !o1.InPort().Stream().Connected(o2.InPort()) {
		t.Error("should be connected")
	}

	if !o2.OutPort().Stream().Connected(o3.InPort()) {
		t.Error("should be connected")
	}

	if !o3.OutPort().Stream().Connected(o4.InPort().Stream().Stream().Stream()) {
		t.Error("should be connected")
	}

	if !o3.OutPort().Connected(o4.InPort().Stream().Stream()) {
		t.Error("should be connected")
	}

	if !o4.OutPort().Stream().Stream().Connected(o1.OutPort().Stream().Stream()) {
		t.Error("should be connected")
	}

	if !o4.OutPort().Stream().Connected(o1.OutPort().Stream()) {
		t.Error("should be connected")
	}

	if !o4.OutPort().Connected(o1.OutPort()) {
		t.Error("should be connected")
	}

	if o2.BasePort() != o1.InPort() {
		t.Error("wrong base port")
	}

	if o3.BasePort() != o2.OutPort() {
		t.Error("wrong base port")
	}

	if !o1.InPort().Connected(o4.InPort()) {
		t.Error("should be connected via base port")
	}

	if !o2.OutPort().Connected(o4.InPort().Stream()) {
		t.Error("should be connected via base port")
	}

	//

	o4.Compile()

	if !o1.InPort().Connected(o1.OutPort()) {
		t.Error("should be connected after merge")
	}

	if !o2.OutPort().Connected(o1.OutPort().Stream()) {
		t.Error("should be connected after merge")
	}

	if !o3.OutPort().Stream().Connected(o5.InPort().Stream()) {
		t.Error("should be connected after merge")
	}

	if !o3.OutPort().Connected(o5.InPort()) {
		t.Error("should be connected after merge")
	}

	//

	o1.OutPort().Stream().Stream().Bufferize()

	go o2.Start()
	go o3.Start()
	go o5.Start()

	o1.InPort().Push(helperJson2I(`[1,2,3]`))
	o1.InPort().Push(helperJson2I(`[]`))
	o1.InPort().Push(helperJson2I(`[4]`))
	assertPortItems(t, helperJson2I(`[[[1],[1,3],[1,3,6]],[],[[1,3,6,10]]]`).([]interface{}), o1.OutPort())
}

func TestNetwork_Maps_Simple(t *testing.T) {
	defIn := helperJson2PortDef(`{"type":"map","map":{"a":{"type":"number"},"b":{"type":"number"}}}`)
	defOut := defIn

	defMap1In := helperJson2PortDef(`{"type":"number"}`)
	defMap1Out := helperJson2PortDef(`{"type":"map","map":{"a":{"type":"number"},"b":{"type":"number"}}}`)

	defMap2In := defMap1Out
	defMap2Out := defMap1In

	evalMap1 := func(in, out *op.Port, store interface{}) {
		for true {
			i := in.Pull()
			if i, ok := i.(float64); ok {
				out.Port("a").Push(2 * i)
				out.Port("b").Push(3 * i)
			} else {
				out.Push(i)
			}
		}
	}

	evalMap2 := func(in, out *op.Port, store interface{}) {
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

	o, _ := op.MakeOperator("", nil, defIn, defOut, nil)
	oMap1, _ := op.MakeOperator("Map1", evalMap1, defMap1In, defMap1Out, o)
	oMap2, _ := op.MakeOperator("Map2", evalMap2, defMap2In, defMap2Out, o)

	o.InPort().Port("a").Connect(oMap2.InPort().Port("a"))
	o.InPort().Port("b").Connect(oMap1.InPort())
	oMap1.OutPort().Port("a").Connect(oMap2.InPort().Port("b"))
	oMap1.OutPort().Port("b").Connect(o.OutPort().Port("b"))
	oMap2.OutPort().Connect(o.OutPort().Port("a"))

	o.OutPort().Port("a").Bufferize()
	o.OutPort().Port("b").Bufferize()

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
		o.InPort().Push(helperJson2I(d))
	}

	assertPortItems(t, helperJson2I(results).([]interface{}), o.OutPort())

}

func TestNetwork_Maps_Complex(t *testing.T) {
	defStrMapStr := helperJson2PortDef(`{"type":"stream","stream":{"type":"map","map":{
		"N":{"type":"stream","stream":{"type":"number"}},
		"n":{"type":"number"},
		"s":{"type":"string"},
		"b":{"type":"boolean"}}}}`)
	defStrMap := helperJson2PortDef(`{"type":"stream","stream":{"type":"map","map":{
		"sum":{"type":"number"},
		"s":{"type":"string"}}}}`)
	defFilterIn := helperJson2PortDef(`{"type":"map","map":{
		"o":{"type":"any"},
		"b":{"type":"boolean"}}}`)
	defFilterOut := helperJson2PortDef(`{"type":"any"}`)
	defAddIn := helperJson2PortDef(`{"type":"map","map":{
		"a":{"type":"number"},
		"b":{"type":"number"}}}`)
	defAddOut := helperJson2PortDef(`{"type":"number"}`)
	defSumIn := helperJson2PortDef(`{"type":"stream","stream":{"type":"number"}}`)
	defSumOut := helperJson2PortDef(`{"type":"number"}`)

	sumEval := func(in, out *op.Port, store interface{}) {
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

	filterEval := func(in, out *op.Port, store interface{}) {
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

	addEval := func(in, out *op.Port, store interface{}) {
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

	o, _ := op.MakeOperator("Global", nil, defStrMapStr, defStrMap, nil)
	sum, _ := op.MakeOperator("Sum", sumEval, defSumIn, defSumOut, o)
	add, _ := op.MakeOperator("Add", addEval, defAddIn, defAddOut, o)
	filter1, _ := op.MakeOperator("Filter1", filterEval, defFilterIn, defFilterOut, o)
	filter2, _ := op.MakeOperator("Filter2", filterEval, defFilterIn, defFilterOut, o)

	o.InPort().Stream().Port("N").Connect(sum.InPort())
	o.InPort().Stream().Port("n").Connect(add.InPort().Port("b"))
	o.InPort().Stream().Port("b").Connect(filter1.InPort().Port("b"))
	o.InPort().Stream().Port("b").Connect(filter2.InPort().Port("b"))
	o.InPort().Stream().Port("s").Connect(filter2.InPort().Port("o"))
	sum.OutPort().Connect(add.InPort().Port("a"))
	add.OutPort().Connect(filter1.InPort().Port("o"))
	filter1.OutPort().Connect(o.OutPort().Stream().Port("sum"))
	filter2.OutPort().Connect(o.OutPort().Stream().Port("s"))

	o.OutPort().Stream().Port("sum").Bufferize()
	o.OutPort().Stream().Port("s").Bufferize()

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
		o.InPort().Push(helperJson2I(d))
	}

	assertPortItems(t, helperJson2I(results).([]interface{}), o.OutPort())
}
