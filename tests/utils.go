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

func assertPortItems(t *testing.T, i []interface{}, p *op.Port) {
	t.Helper()
	for _, e := range i {
		a := p.Pull()
		if !reflect.DeepEqual(e, a) {
			fmt.Errorf("wrong value:\nexpected: %#v,\nactual:   %#v", e, a)
			break
		}
	}
}
