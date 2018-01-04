package builtin

import (
	"slang/op"
	"slang/tests"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOperatorCreator_Fork_IsRegistered(t *testing.T) {
	a := assert.New(t)

	ocFork := getCreatorFunc("fork")
	a.NotNil(ocFork)
}

func TestBuiltin_OperatorFork__InPorts(t *testing.T) {
	a := assert.New(t)

	o, err := getCreatorFunc("fork")(op.InstanceDef{Operator: "fork"}, nil)
	a.NoError(err)

	a.NotNil(o.In().Map("i"))
	a.NotNil(o.In().Map("select"))
}

func TestBuiltin_OperatorFork__OutPorts(t *testing.T) {
	a := assert.New(t)

	o, err := getCreatorFunc("fork")(op.InstanceDef{Operator: "fork"}, nil)
	a.NoError(err)

	a.NotNil(o.Out().Map("true"))
	a.NotNil(o.Out().Map("false"))
}

func TestBuiltin_OperatorFork__Correct(t *testing.T) {
	a := assert.New(t)

	o, err := getCreatorFunc("fork")(op.InstanceDef{Operator: "fork"}, nil)
	a.NoError(err)

	o.Out().Map("true").Bufferize()
	o.Out().Map("false").Bufferize()
	o.Start()

	datIn := map[string][]interface{}{
		"i":      []interface{}{"hallo", "welt", 100, 200},
		"select": []interface{}{true, false, true, false},
	}

	for _, i := range datIn["i"] {
		o.In().Map("i").Push(i)
	}
	for _, i := range datIn["select"] {
		o.In().Map("select").Push(i)
	}

	tests.AssertPortItems(t, []interface{}{"hallo", 100}, o.Out().Map("true"))
	tests.AssertPortItems(t, []interface{}{"welt", 200}, o.Out().Map("false"))
}
