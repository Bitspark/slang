package elem

import (
	"github.com/Bitspark/slang/tests/assertions"
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/Bitspark/slang/pkg/core"
)

// Test if fork operator is registered under the correct name
func Test_ElemCtrl_Loop_CreatorFuncIsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocLoop := getBuiltinCfg("slang.stream.Loop")
	a.NotNil(ocLoop)
}

// Test if the signature is correct
func Test_ElemCtrl_Loop__Signature(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	lop, err := buildOperator(
		core.InstanceDef{
			Name:     "loop",
			Operator: "slang.stream.Loop",
			Generics: map[string]*core.TypeDef{
				"stateType": {
					Type: "number",
				},
				"itemType": {
					Type: "number",
				},
			},
		},
	)

	r.NoError(err)
	r.NotNil(lop)

	// In port
	in := lop.Main().In()
	r.NotNil(in)

	// Out port
	out := lop.Main().Out()
	r.NotNil(out)
	r.Equal(core.TYPE_MAP, out.Type())
	r.NotNil(out.Map("result"))
	r.NotNil(out.Map("items"))
	r.Equal(core.TYPE_STREAM, out.Map("items").Type())

	// Iterator delegate
	dlgIter := lop.Delegate("iterator")
	r.NotNil(dlgIter)

	// Delegate out port
	dlgIterOut := dlgIter.Out()
	r.NotNil(dlgIterOut)

	// Delegate in port
	dlgIterIn := dlgIter.In()
	r.NotNil(dlgIterIn)
	r.Equal(core.TYPE_MAP, dlgIterIn.Type())
	r.NotNil(dlgIterIn.Map("state"))
	r.NotNil(dlgIterIn.Map("item"))

	// Controller delegate
	dlgCtrl := lop.Delegate("controller")
	r.NotNil(dlgCtrl)

	// Delegate out port
	dlgCtrlOut := dlgCtrl.Out()
	r.NotNil(dlgCtrlOut)

	// Delegate in port
	dlgCtrlIn := dlgCtrl.In()
	r.NotNil(dlgCtrlIn)
	a.Equal(core.TYPE_BOOLEAN, dlgCtrlIn.Type())
}

// Test if generics are replaced correctly
func Test_ElemCtrl_Loop__GenericType(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	types := []string{"number", "boolean", "string"}
	typesInt := []int{core.TYPE_NUMBER, core.TYPE_BOOLEAN, core.TYPE_STRING}

	for i, stateTp := range types {
		for j, itemTp := range types {
			stateTpi := typesInt[i]
			itemTpi := typesInt[j]

			lop, err := buildOperator(
				core.InstanceDef{
					Name:     "fork",
					Operator: "slang.stream.Loop",
					Generics: map[string]*core.TypeDef{
						"stateType": {
							Type: stateTp,
						},
						"itemType": {
							Type: itemTp,
						},
					},
				},
			)

			r.NoError(err)
			r.NotNil(lop)

			// In port
			in := lop.Main().In()
			r.NotNil(in)
			a.Equal(stateTpi, in.Type())

			// Out port
			out := lop.Main().Out()
			r.NotNil(out)
			a.Equal(stateTpi, out.Map("result").Type())
			a.Equal(itemTpi, out.Map("items").Stream().Type())

			// Iterator delegate
			dlgIter := lop.Delegate("iterator")

			// Delegate out port
			dlgIterOut := dlgIter.Out()
			a.Equal(stateTpi, dlgIterOut.Type())

			// Delegate in port
			dlgIterIn := dlgIter.In()
			a.Equal(stateTpi, dlgIterIn.Map("state").Type())
			a.Equal(itemTpi, dlgIterIn.Map("item").Type())

			// Controller delegate
			dlgCtrl := lop.Delegate("controller")

			// Delegate out port
			dlgCtrlOut := dlgCtrl.Out()
			a.Equal(stateTpi, dlgCtrlOut.Type())
		}
	}
}

// Test if generics are replaced correctly
func Test_ElemCtrl_Loop__Behavior(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	lop, err := buildOperator(
		core.InstanceDef{
			Name:     "loop",
			Operator: "slang.stream.Loop",
			Generics: map[string]*core.TypeDef{
				"stateType": {
					Type: "number",
				},
				"itemType": {
					Type: "number",
				},
			},
		},
	)

	r.NoError(err)
	r.NotNil(lop)

	dlgIter := lop.Delegate("iterator")
	dlgCtrl := lop.Delegate("controller")

	// Bufferize
	lop.Main().Out().Bufferize()
	dlgIter.Out().Bufferize()
	dlgCtrl.Out().Bufferize()

	lop.Start()

	// Values we push into the loop at its in port
	inVals := []interface{}{
		0.0,
		10.0,
		100.0,
	}

	// Values we push into the loop as controller (indicating when to stop, hence 'false' is always last)
	ctrlVals := []interface{}{
		[]interface{}{false},
		[]interface{}{true, false},
		[]interface{}{true, true, true, false},
	}

	// These values are expected to arrive at the iterator's in port
	// We also push these values as new states to simulate new state calculation of the iterator
	stateVals := []interface{}{
		[]interface{}{
			map[string]interface{}{"state": 0.0, "item": 0.0},
		},
		[]interface{}{
			map[string]interface{}{"state": 10.0, "item": 0.0},
			map[string]interface{}{"state": 11.0, "item": -11.0},
		},
		[]interface{}{
			map[string]interface{}{"state": 100.0, "item": 0.0},
			map[string]interface{}{"state": 101.0, "item": -101.0},
			map[string]interface{}{"state": 102.0, "item": -102.0},
			map[string]interface{}{"state": 103.0, "item": -103.0},
		},
	}

	// These values are expected to be emitted by the loop
	expectedVals := []interface{}{
		map[string]interface{}{
			"result": 0.0,
			"items":  []interface{}{},
		},
		map[string]interface{}{
			"result": 11.0,
			"items":  []interface{}{-11.0},
		},
		map[string]interface{}{
			"result": 103.0,
			"items":  []interface{}{-101.0, -102.0, -103.0},
		},
	}

	for i, in := range inVals {
		// Push values
		lop.Main().In().Push(in)

		// Test output behavior
		for j, ctrlVal := range ctrlVals[i].([]interface{}) {
			a.Equal(stateVals[i].([]interface{})[j].(map[string]interface{})["state"], dlgCtrl.Out().Pull())
			dlgCtrl.In().Push(ctrlVal)

			if ctrlVal.(bool) {
				a.Equal(stateVals[i].([]interface{})[j].(map[string]interface{})["state"], dlgIter.Out().Pull())
				dlgIter.In().Push(stateVals[i].([]interface{})[j+1])
			}
		}

		a.Equal(expectedVals[i], lop.Main().Out().Pull())
	}
}
