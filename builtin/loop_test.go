package builtin

import (
	"github.com/stretchr/testify/require"
	"slang/core"
	"slang/tests/assertions"
	"testing"
)

func TestOperatorCreator_Loop_IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocLoop := getBuiltinCfg("loop")
	a.NotNil(ocLoop)
}

func TestBuiltin_Loop__Simple(t *testing.T) {
	a := assertions.New(t)
	lo, err := MakeOperator(
		core.InstanceDef{
			Operator: "loop",
			Generics: map[string]*core.PortDef{
				"stateType": {
					Type: "number",
				},
			},
		},
	)
	a.NoError(err)
	a.NotNil(lo)

	// Condition operator
	co, _ := core.NewOperator("cond", func(in, out *core.Port, store interface{}) {
		for true {
			i := in.Pull()
			f, ok := i.(float64)
			if !ok {
				out.Push(i)
			} else {
				out.Push(f < 10.0)
			}
		}
	}, core.PortDef{Type: "number"}, core.PortDef{Type: "boolean"})

	// Double function operator
	fo, _ := core.NewOperator("double", func(in, out *core.Port, store interface{}) {
		for true {
			i := in.Pull()
			f, ok := i.(float64)
			if !ok {
				out.Push(i)
			} else {
				out.Push(f * 2.0)
			}
		}
	}, core.PortDef{Type: "number"}, core.PortDef{Type: "number"})

	// Connect
	a.NoError(lo.Out().Map("state").Stream().Connect(fo.In()))
	a.NoError(lo.Out().Map("state").Stream().Connect(co.In()))
	a.NoError(fo.Out().Connect(lo.In().Map("iteration").Stream().Map("state")))
	a.NoError(co.Out().Connect(lo.In().Map("iteration").Stream().Map("continue")))

	lo.Out().Bufferize()

	lo.In().Map("init").Push(1.0)
	lo.In().Map("init").Push(10.0)

	lo.Start()
	fo.Start()
	co.Start()

	a.PortPushes([]interface{}{16.0, 10.0}, lo.Out().Map("end"))
}

func TestBuiltin_Loop__Fibo(t *testing.T) {
	a := assertions.New(t)
	stateType := core.PortDef{
		Type: "map",
		Map: map[string]*core.PortDef{
			"i":      {Type: "number"},
			"fib":    {Type: "number"},
			"oldFib": {Type: "number"},
		},
	}
	lo, err := MakeOperator(
		core.InstanceDef{
			Operator: "loop",
			Generics: map[string]*core.PortDef{
				"stateType": &stateType,
			},
		},
	)
	require.NoError(t, err)
	a.NotNil(lo)
	require.Equal(t, core.TYPE_MAP, lo.In().Map("init").Type())
	require.Equal(t, core.TYPE_NUMBER, lo.In().Map("init").Map("i").Type())

	// Condition operator
	co, _ := core.NewOperator("cond", func(in, out *core.Port, store interface{}) {
		for true {
			i := in.Pull()
			fm, ok := i.(map[string]interface{})
			if !ok {
				out.Push(i)
			} else {
				i := fm["i"].(float64)
				out.Push(i > 0.0)
			}
		}
	}, stateType, core.PortDef{Type: "boolean"})

	// Fibonacci function operator
	fo, _ := core.NewOperator("fib", func(in, out *core.Port, store interface{}) {
		for true {
			i := in.Pull()
			fm, ok := i.(map[string]interface{})
			if !ok {
				out.Push(i)
			} else {
				i := fm["i"].(float64) - 1
				oldFib := fm["fib"].(float64)
				fib := fm["oldFib"].(float64) + oldFib
				out.Push(map[string]interface{}{"i": i, "fib": fib, "oldFib": oldFib})
			}
		}
	}, stateType, stateType)

	// Connect
	a.NoError(lo.Out().Map("state").Stream().Connect(fo.In()))
	a.NoError(lo.Out().Map("state").Stream().Connect(co.In()))
	a.NoError(fo.Out().Connect(lo.In().Map("iteration").Stream().Map("state")))
	a.NoError(co.Out().Connect(lo.In().Map("iteration").Stream().Map("continue")))

	lo.Out().Bufferize()

	lo.In().Map("init").Push(map[string]interface{}{"i": 10.0, "fib": 1.0, "oldFib": 0.0})
	lo.In().Map("init").Push(map[string]interface{}{"i": 20.0, "fib": 1.0, "oldFib": 0.0})

	lo.Start()
	fo.Start()
	co.Start()

	a.PortPushes([]interface{}{
		map[string]interface{}{"i": 0.0, "fib": 89.0, "oldFib": 55.0},
		map[string]interface{}{"i": 0.0, "fib": 10946.0, "oldFib": 6765.0},
	}, lo.Out().Map("end"))
}

func TestBuiltin_Loop__MarkersPushedCorrectly(t *testing.T) {
	a := assertions.New(t)
	lo, err := MakeOperator(
		core.InstanceDef{
			Operator: "loop",
			Generics: map[string]*core.PortDef{
				"stateType": {
					Type: "number",
				},
			},
		},
	)
	a.NoError(err)
	a.NotNil(lo)

	lo.Out().Bufferize()

	lo.Start()

	pInit := lo.In().Map("init")
	pIteration := lo.In().Map("iteration")
	pState := lo.Out().Map("state").Stream()
	pEnd := lo.Out().Map("end")

	bos := core.BOS{}
	pInit.Push(bos)
	a.Nil(pEnd.Poll())
	a.Equal(bos, pState.Pull())

	pIteration.Push(bos)

	a.Equal(bos, pEnd.Pull())
	a.Nil(pState.Poll())

	eos := core.BOS{}
	pInit.Push(eos)
	a.Nil(pEnd.Poll())
	a.Equal(eos, pState.Pull())

	pIteration.Push(eos)

	a.Equal(eos, pEnd.Pull())
	a.Nil(pState.Poll())
}
