package main

import (
	"encoding/json"
	"testing"
	"reflect"
	"fmt"
)

func helperJson2I(str string) interface{} {
	var obj interface{}
	json.Unmarshal([]byte(str), &obj)
	return obj
}

func helperJson2Map(str string) map[string]interface{} {
	m, _ := helperJson2I(str).(map[string]interface{})
	return m
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

func assertPortItems(t *testing.T, i []interface{}, p *Port) {
	t.Helper()
	for _, e := range i {
		a := p.Pull()
		if !reflect.DeepEqual(e, a) {
			t.Error(fmt.Sprintf("wrong value:\nexpected: %#v,\nactual:   %#v", e, a))
			break
		}
	}
}