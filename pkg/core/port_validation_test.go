package core

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func validationOperator(t *testing.T, name string, function OFunc, typ TypeDef) *Operator {
	t.Helper()
	o, err := NewOperator(name, function, nil, nil, nil, Blueprint{
		ServiceDefs: map[string]*ServiceDef{MAIN_SERVICE: {In: typ, Out: typ}},
	})
	require.NoError(t, err)
	return o
}

func TestFullyConnectedReportsAllNestedMissingPorts(t *testing.T) {
	typ := TypeDef{Type: "map", Map: TypeDefMap{
		"connected": {Type: "number"},
		"nested":    {Type: "map", Map: TypeDefMap{"missing": {Type: "string"}}},
		"rows": {Type: "stream", Stream: &TypeDef{Type: "map", Map: TypeDefMap{
			"a": {Type: "number"}, "b": {Type: "number"},
		}}},
	}}
	o := validationOperator(t, "blueprint", nil, typ)
	require.NoError(t, o.Main().In().Map("connected").Connect(o.Main().Out().Map("connected")))
	for i := 0; i < 10; i++ {
		require.EqualError(t, o.Main().Out().FullyConnected(),
			"unconnected output ports: blueprint)nested.missing, blueprint)rows.~.a, blueprint)rows.~.b")
	}
}

func TestFullyConnectedFollowsNestedForwarding(t *testing.T) {
	typ := TypeDef{Type: "stream", Stream: &TypeDef{Type: "number"}}
	root := validationOperator(t, "root", nil, typ)
	child := validationOperator(t, "child", nil, typ)
	child.SetParent(root)
	require.NoError(t, child.Main().Out().Connect(root.Main().Out()))
	require.Error(t, root.Main().Out().FullyConnected())
	require.NoError(t, child.Main().In().Connect(child.Main().Out()))
	// A child input is not an external source until it is wired to the root.
	require.Error(t, root.Main().Out().FullyConnected())
	require.NoError(t, root.Main().In().Connect(child.Main().In()))
	require.NoError(t, root.Main().Out().FullyConnected())
	root.Compile()
	require.NoError(t, root.Main().Out().FullyConnected())
}

func TestFullyConnectedAcceptsBuiltinProducer(t *testing.T) {
	typ := TypeDef{Type: "number"}
	root := validationOperator(t, "root", nil, typ)
	producer := validationOperator(t, "producer", func(*Operator) {}, typ)
	producer.SetParent(root)
	require.NoError(t, producer.Main().Out().FullyConnected())
	require.NoError(t, producer.Main().Out().Connect(root.Main().Out()))
	require.NoError(t, root.Main().Out().FullyConnected())
}

func TestFullyConnectedRejectsSourceCycle(t *testing.T) {
	typ := TypeDef{Type: "number"}
	root := validationOperator(t, "root", nil, typ)
	a := validationOperator(t, "a", nil, typ)
	b := validationOperator(t, "b", nil, typ)
	a.SetParent(root)
	b.SetParent(root)
	require.NoError(t, a.Main().In().Connect(a.Main().Out()))
	require.NoError(t, b.Main().In().Connect(b.Main().Out()))
	require.NoError(t, a.Main().Out().Connect(b.Main().In()))
	require.NoError(t, b.Main().Out().Connect(a.Main().In()))
	require.NoError(t, a.Main().Out().Connect(root.Main().Out()))
	require.EqualError(t, root.Main().Out().FullyConnected(), "unconnected output ports: root)")
}

func TestFullyConnectedAcceptsEmptyMap(t *testing.T) {
	o := validationOperator(t, "empty", nil, TypeDef{Type: "map", Map: TypeDefMap{}})
	require.NoError(t, o.Main().Out().FullyConnected())
}
