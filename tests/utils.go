package tests

import (
	"encoding/json"
	"github.com/Bitspark/slang"
	"github.com/Bitspark/slang/core"
)

func parseJSON(str string) interface{} {
	var obj interface{}
	json.Unmarshal([]byte(str), &obj)
	return obj
}

func validateJSONOperatorDef(jsonDef string) (core.OperatorDef, error) {
	def, _ := slang.ParseJSONOperatorDef(jsonDef)
	return def, def.Validate()
}

func validateJSONInstanceDef(jsonDef string) (core.InstanceDef, error) {
	def := core.InstanceDef{}
	json.Unmarshal([]byte(jsonDef), &def)
	return def, def.Validate()
}
