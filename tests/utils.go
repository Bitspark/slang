package tests

import (
	"encoding/json"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/api"
)

func parseJSON(str string) interface{} {
	var obj interface{}
	json.Unmarshal([]byte(str), &obj)
	return obj
}

func validateJSONOperatorDef(jsonDef string) (core.OperatorDef, error) {
	def, _ := api.ParseJSONOperatorDef(jsonDef)
	return def, def.Validate()
}

func validateJSONInstanceDef(jsonDef string) (core.InstanceDef, error) {
	def := core.InstanceDef{}
	json.Unmarshal([]byte(jsonDef), &def)
	return def, def.Validate()
}
