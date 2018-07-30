package elem

import (
	"github.com/Bitspark/slang/tests/assertions"
	"testing"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/stretchr/testify/require"
)

// Test if fork operator is registered under the correct name
func Test_ElemCtrl_Fork_CreatorFuncIsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFork := getBuiltinCfg("slang.control.fork")
	a.NotNil(ocFork)
}

// Test if the signature is correct
func Test_ElemCtrl_Fork__Signature(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	fop, err := buildOperator(
		core.InstanceDef{
			Name:     "fork",
			Operator: "slang.control.fork",
			Generics: map[string]*core.TypeDef{
				"itemType": {
					Type: "number",
				},
			},
		},
	)

	r.NoError(err)
	r.NotNil(fop)

	// In port
	in := fop.Main().In()
	r.NotNil(in)

	// Out port
	out := fop.Main().Out()
	r.NotNil(out)
	r.Equal(core.TYPE_MAP, out.Type())
	r.NotNil(out.Map("true"))
	r.NotNil(out.Map("false"))
	r.NotNil(out.Map("control"))
	r.Equal(core.TYPE_STREAM, out.Map("true").Type())
	r.Equal(core.TYPE_STREAM, out.Map("false").Type())
	r.Equal(core.TYPE_STREAM, out.Map("control").Type())
	a.Equal(core.TYPE_BOOLEAN, out.Map("control").Stream().Type())

	// Delegate
	dlg := fop.Delegate("controller")
	r.NotNil(dlg)

	// Delegate out port
	dlgOut := dlg.Out()
	r.NotNil(dlgOut)

	// Delegate in port
	dlgIn := dlg.In()
	r.NotNil(dlgIn)
	a.Equal(core.TYPE_BOOLEAN, dlgIn.Type())
}

// Test if generics are replaced correctly
func Test_ElemCtrl_Fork__GenericType(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	types := []string{"number", "boolean", "string"}
	typesInt := []int{core.TYPE_NUMBER, core.TYPE_BOOLEAN, core.TYPE_STRING}

	for i, tp := range types {
		tpi := typesInt[i]
		fop, err := buildOperator(
			core.InstanceDef{
				Name:     "fork",
				Operator: "slang.control.fork",
				Generics: map[string]*core.TypeDef{
					"itemType": {
						Type: tp,
					},
				},
			},
		)

		r.NoError(err)
		r.NotNil(fop)

		// In port
		in := fop.Main().In()
		a.Equal(tpi, in.Stream().Type())

		// Out port
		out := fop.Main().Out()
		a.Equal(tpi, out.Map("true").Stream().Type())
		a.Equal(tpi, out.Map("false").Stream().Type())

		// Delegate
		dlg := fop.Delegate("controller")

		// Delegate out port
		dlgOut := dlg.Out()
		a.Equal(tpi, dlgOut.Type())
	}
}

// Test if fork operator redirects items correctly to true and false ports and emits correct values for control port
func Test_ElemCtrl_Fork__Forking(t *testing.T) {
	a := assertions.New(t)
	r := require.New(t)

	fop, err := buildOperator(
		core.InstanceDef{
			Name:     "fork",
			Operator: "slang.control.fork",
			Generics: map[string]*core.TypeDef{
				"itemType": {
					Type: "number",
				},
			},
		},
	)

	r.NoError(err)
	r.NotNil(fop)

	dlg := fop.Delegate("controller")

	// Set stream source for delegate in port
	dlg.In().SetStreamSource(dlg.Out())

	// Bufferize
	fop.Main().Out().Bufferize()
	dlg.Out().Bufferize()

	fop.Start()

	items := []interface{}{1.0, 2.0, 3.0, 4.0, 5.0}
	control := []interface{}{true, false, false, true, true}
	fop.Main().In().Push(items)

	for idx, item := range items {
		i := dlg.Out().Pull()
		a.Equal(item, i)
		dlg.In().Push(control[idx])
	}

	trueItems := fop.Main().Out().Map("true").Pull()
	falseItems := fop.Main().Out().Map("false").Pull()
	controlValues := fop.Main().Out().Map("control").Pull()

	a.Equal([]interface{}{1.0, 4.0, 5.0}, trueItems)
	a.Equal([]interface{}{2.0, 3.0}, falseItems)
	a.Equal(control, controlValues)
}

// TODO: Test stream source connection