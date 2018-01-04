package slang

import (
	"encoding/json"
	"slang/op"
)

func getJSONOperatorDef(jsonDef string) *op.OperatorDef {
	def := &op.OperatorDef{}
	json.Unmarshal([]byte(jsonDef), def)
	return def
}
