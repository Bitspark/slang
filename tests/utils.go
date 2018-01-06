package tests

import (
	"encoding/json"
	"slang"
	"slang/op"
)

func parseJSON(str string) interface{} {
	var obj interface{}
	json.Unmarshal([]byte(str), &obj)
	return obj
}

func validateJSONOperatorDef(jsonDef string) (op.OperatorDef, error) {
	def := slang.ParseOperatorDef(jsonDef)
	return def, def.Validate()
}

func validateJSONInstanceDef(jsonDef string) (op.InstanceDef, error) {
	def := op.InstanceDef{}
	json.Unmarshal([]byte(jsonDef), &def)
	return def, def.Validate()
}
