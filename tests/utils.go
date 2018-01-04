package tests

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slang/op"
	"testing"
)

func parseJSON(str string) interface{} {
	var obj interface{}
	json.Unmarshal([]byte(str), &obj)
	return obj
}

func validateJSONOperatorDef(jsonDef string) (op.OperatorDef, error) {
	def := op.ParseOperatorDef(jsonDef)
	return def, def.Validate()
}

func validateJSONInstanceDef(jsonDef string) (op.InstanceDef, error) {
	def := op.InstanceDef{}
	json.Unmarshal([]byte(jsonDef), &def)
	return def, def.Validate()
}

func assertTrue(t *testing.T, b bool) {
	if !b {
		t.Error("expected to be true")
	}
}

func assertFalse(t *testing.T, b bool) {
	if b {
		t.Error("expected to be false")
	}
}

func assertPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	f()
}

func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Error("expected error")
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Error(err.Error())
	}
}

func assertNil(t *testing.T, a interface{}) {
	t.Helper()
	if a != nil && !reflect.ValueOf(a).IsNil() {
		t.Error("instance should be nil")
	}
}

func assertNotNil(t *testing.T, a interface{}) {
	t.Helper()
	if a == nil || reflect.ValueOf(a).IsNil() {
		t.Error("instance is nil")
	}
}

func assertPortItems(t *testing.T, i []interface{}, p *op.Port) {
	t.Helper()
	for _, e := range i {
		a := p.Pull()
		if !reflect.DeepEqual(e, a) {
			t.Error(fmt.Sprintf("wrong value:\nexpected: %#v,\nactual:   %#v", e, a))
			break
		}
	}
}
