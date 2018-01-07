package builtin

import (
	"testing"
	"slang/tests/assertions"
	"slang/core"
)

func TestOperatorCreator_Loop_IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocLoop := getBuiltinCfg("loop")
	a.NotNil(ocLoop)
}

func TestBuiltin_Loop__SimpleLoop(t *testing.T) {
	a := assertions.New(t)
	lo, err := MakeOperator(&core.InstanceDef{Operator: "loop"})
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

	lo.Out().Map("end").Bufferize()

	lo.In().Map("init").Push(1.0)
	lo.In().Map("init").Push(10.0)

	lo.Start()
	fo.Start()
	co.Start()

	a.PortPushes([]interface{}{16.0, 10.0}, lo.Out().Map("end"))
}

func TestBuiltin_Loop__FibLoop(t *testing.T) {
	a := assertions.New(t)
	lo, err := MakeOperator(&core.InstanceDef{Operator: "loop"})
	a.NoError(err)
	a.NotNil(lo)

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
	}, core.PortDef{Type: "primitive"}, core.PortDef{Type: "boolean"})

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
	}, core.PortDef{Type: "primitive"}, core.PortDef{Type: "primitive"})

	// Connect
	a.NoError(lo.Out().Map("state").Stream().Connect(fo.In()))
	a.NoError(lo.Out().Map("state").Stream().Connect(co.In()))
	a.NoError(fo.Out().Connect(lo.In().Map("iteration").Stream().Map("state")))
	a.NoError(co.Out().Connect(lo.In().Map("iteration").Stream().Map("continue")))

	lo.Out().Map("end").Bufferize()

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
