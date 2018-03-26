package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
	"strings"
	"encoding/csv"
	"io"
	"errors"
)

type csvReadStore struct {
	colMapping []string
	delimiter  rune
}

func checkMapping(port *core.Port, mapping []string) error {
	for i := 0; i < len(mapping); i++ {
		for j := i + 1; j < len(mapping); j++ {
			if mapping[i] == mapping[j] {
				return errors.New("duplicate " + mapping[i])
			}
		}
		if port.Map(mapping[i]) == nil {
			return errors.New("no map for " + mapping[i])
		}
	}
	return nil
}

var csvReadOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		Services: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.PortDef{
					Type: "string",
				},
				Out: core.PortDef{
					Type: "stream",
					Stream: &core.PortDef{
						Type:    "generic",
						Generic: "colMap",
					},
				},
			},
		},
		Delegates: map[string]*core.DelegateDef{
		},
	},
	oFunc: func(srvs map[string]*core.Service, dels map[string]*core.Delegate, store interface{}) {
		in := srvs[core.MAIN_SERVICE].In()
		out := srvs[core.MAIN_SERVICE].Out()
		for {
			csvText, marker := in.PullString()
			if marker != nil {
				out.Push(marker)
				continue
			}

			outStream := out.Stream()

			out.PushBOS()

			csvProps := store.(csvReadStore)
			mapping := csvProps.colMapping
			mapSize := outStream.MapSize()

			r := csv.NewReader(strings.NewReader(csvText))
			r.Comma = csvProps.delimiter
			for {
				rec, err := r.Read()
				if err == io.EOF {
					break
				}
				if len(rec) != mapSize {
					break
				}

				if mapping == nil {
					if err := checkMapping(outStream, rec); err != nil {
						break
					}
					mapping = rec
				} else {
					for i, col := range mapping {
						outStream.Map(col).Push(rec[i])
					}
				}
			}
			out.PushEOS()
		}
	},
	oPropFunc: func(op *core.Operator, i map[string]interface{}) error {
		csvParams := csvReadStore{}

		if delim, ok := i["delimiter"]; ok {
			delimStr := delim.(string)
			if len(delimStr) != 1 {
				return errors.New("delimiter must not be single character")
			}
			csvParams.delimiter = rune(delimStr[0])
		} else {
			csvParams.delimiter = ','
		}

		if mapping, ok := i["colMapping"]; ok {
			mapArray := mapping.([]interface{})
			csvParams.colMapping = make([]string, len(mapArray))
			for i, row := range mapArray {
				csvParams.colMapping[i] = row.(string)
			}
			if err := checkMapping(op.Main().Out().Stream(), csvParams.colMapping); err != nil {
				return err
			}
		}

		op.SetStore(csvParams)

		return nil
	},
	oConnFunc: func(dest, src *core.Port) error {
		return nil
	},
}
