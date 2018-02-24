package builtin

import (
	"github.com/stretchr/testify/require"
	"slang/core"
	"slang/tests/assertions"
	"testing"
)

func TestOperatorCreator_Loop_IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocLoop := getBuiltinCfg("slang.loop")
	a.NotNil(ocLoop)
}

func TestBuiltin_Loop__Simple(t *testing.T) {
	a := assertions.New(t)
	lo, err := MakeOperator(
		core.InstanceDef{
			Name: "loop",
			Operator: "slang.loop",
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
	co, _ := core.NewOperator("cond", func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
		for true {
			i := in.Pull()
			f, ok := i.(float64)
			if !ok {
				out.Push(i)
			} else {
				out.Push(f < 10.0)
			}
		}
	}, core.PortDef{Type: "number"}, core.PortDef{Type: "boolean"},
		nil)

	// Double function operator
	fo, _ := core.NewOperator("double", func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
		for true {
			i := in.Pull()
			f, ok := i.(float64)
			if !ok {
				out.Push(i)
			} else {
				out.Push(f * 2.0)
			}
		}
	}, core.PortDef{Type: "number"}, core.PortDef{Type: "number"},
		nil)

	// Connect
	a.NoError(lo.Delegate("iteration").Out().Stream().Connect(fo.In()))
	a.NoError(lo.Delegate("iteration").Out().Stream().Connect(co.In()))
	a.NoError(fo.Out().Connect(lo.Delegate("iteration").In().Stream().Map("state")))
	a.NoError(co.Out().Connect(lo.Delegate("iteration").In().Stream().Map("continue")))

	lo.Out().Bufferize()

	lo.In().Push(1.0)
	lo.In().Push(10.0)

	lo.Start()
	fo.Start()
	co.Start()

	a.PortPushes([]interface{}{16.0, 10.0}, lo.Out())
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
			Operator: "slang.loop",
			Generics: map[string]*core.PortDef{
				"stateType": &stateType,
			},
		},
	)
	require.NoError(t, err)
	a.NotNil(lo)
	require.Equal(t, core.TYPE_MAP, lo.In().Type())
	require.Equal(t, core.TYPE_NUMBER, lo.In().Map("i").Type())

	// Condition operator
	co, _ := core.NewOperator("cond", func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
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
	}, stateType, core.PortDef{Type: "boolean"},
		nil)

	// Fibonacci function operator
	fo, _ := core.NewOperator("fib", func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
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
	}, stateType, stateType,
		nil)

	// Connect
	a.NoError(lo.Delegate("iteration").Out().Stream().Connect(fo.In()))
	a.NoError(lo.Delegate("iteration").Out().Stream().Connect(co.In()))
	a.NoError(fo.Out().Connect(lo.Delegate("iteration").In().Stream().Map("state")))
	a.NoError(co.Out().Connect(lo.Delegate("iteration").In().Stream().Map("continue")))

	lo.Out().Bufferize()

	lo.In().Push(map[string]interface{}{"i": 10.0, "fib": 1.0, "oldFib": 0.0})
	lo.In().Push(map[string]interface{}{"i": 20.0, "fib": 1.0, "oldFib": 0.0})

	lo.Start()
	fo.Start()
	co.Start()

	a.PortPushes([]interface{}{
		map[string]interface{}{"i": 0.0, "fib": 89.0, "oldFib": 55.0},
		map[string]interface{}{"i": 0.0, "fib": 10946.0, "oldFib": 6765.0},
	}, lo.Out())
}

func TestBuiltin_Loop__MarkersPushedCorrectly(t *testing.T) {
	a := assertions.New(t)
	lo, err := MakeOperator(
		core.InstanceDef{
			Operator: "slang.loop",
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
	lo.Delegate("iteration").Out().Bufferize()

	lo.Start()

	pInit := lo.In()
	pIteration := lo.Delegate("iteration").In()
	pState := lo.Delegate("iteration").Out().Stream()
	pEnd := lo.Out()

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
