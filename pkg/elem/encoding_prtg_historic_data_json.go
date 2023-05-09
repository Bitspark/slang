package elem

import (
	"encoding/json"
	"strings"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/log"
	"github.com/google/uuid"
	"github.com/thoas/go-funk"
)

/*

Parses PRTG Historic Data payload

{
  "prtg-version": "23.1.82.2175",
  "treesize": 1200,
  "histdata": [
    {
      "datetime": "27.03.2023 00:00:00",
      " Power from PV ": "",
      " Solar Energy Today ": "",
      "coverage": "0 %"
    },
	...
	{
      "datetime": "27.03.2023 12:37:00",
      " Power from PV ": 22.6910,
      " Solar Energy Today ": 112.5000,
      "coverage": "100 %"
    },
    {
      "datetime": "27.03.2023 12:38:00",
      " Power from PV ": 21.4000,
      " Solar Energy Today ": 112.9000,
      "coverage": "100 %"
    },
    {
      "datetime": "27.03.2023 12:39:00",
      " Power from PV ": 21.8630,
      " Solar Energy Today ": 113.2000,
      "coverage": "100 %"
    }
  ]
}

into Stream of map

[{"Power from PV": 0, "Solar Energy Today": 0}, ...]
*/
type HistData struct {
	DateTime string                 `json:"datetime"`
	Coverage string                 `json:"coverage"`
	Fields   map[string]interface{} `json:"-"`
}

type PRTGData struct {
	PRTGVersion string     `json:"prtg-version"`
	TreeSize    int        `json:"treesize"`
	HistData    []HistData `json:"histdata"`
}


func (h *HistData) UnmarshalJSON(data []byte) error {
	h.Fields = make(map[string]interface{})
	return json.Unmarshal(data, &h.Fields)
}

// XXX Property To Portname --> sanitize property so it can be used as port names
func getChannels(channelNames []interface{}, prtgData *PRTGData) []map[string]float64 {
	histdata := prtgData.HistData

	channelStream := make([]map[string]float64, 0)

	for i := range histdata {

		if i == 0 {
			continue
		}

		d := histdata[i].Fields

		channelValues := make(map[string]float64)

		for _, channelName := range channelNames {
			for k, v := range d {
				k = strings.TrimSpace(k)
				if k == channelName {
					if channelValue, ok := v.(float64); ok {
						channelValues[channelName.(string)] = channelValue
					}
				}
			}
		}
		channelStream = append(channelStream, channelValues)

	}

	return channelStream
}

var encodingPRTGHistDataId = uuid.MustParse("71f01aee-860a-49a4-8fe0-2a449301ea54")
var encodingPRTGHistDataCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: encodingPRTGHistDataId,
		Meta: core.BlueprintMetaDef{
			Name:             "decode PRTG API histoic data",
			ShortDescription: "decodes PRTG API historic data and pass channel values",
			Icon:             "brackets-curly",
			Tags:             []string{"encoding"},
			DocURL:           "https://bitspark.de/slang/docs/operator/decode-json",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "binary",
				},
				Out: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"valid": {
							Type: "boolean",
						},
						"channels": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type: "map",
								Map: map[string]*core.TypeDef{
									"{ChannelNames}": {
										Type:    "number",
									},
								},
							}, 
						},
					},
				},
			},
		},
		PropertyDefs: core.PropertyMap{
			"ChannelNames": {
				Type: "stream",
				Stream: &core.TypeDef{
					Type: "string",
				},
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()

		channelNames := op.Property("ChannelNames").([]interface{})

		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			// Unmarshal JSON data
			var prtgData PRTGData
			if err := json.Unmarshal([]byte(i.(core.Binary)), &prtgData); err != nil {
				log.Error("cannot parse PRTG historic data payload:", prtgData)
				out.Map("item").Push(nil)
				out.Map("valid").Push(false)
				continue
			}

			channelValues := getChannels(channelNames, &prtgData)
			invalid := false
			
			// check if channelValues provides values for all expected channels
			if len(channelValues) > 0 {
				providedChannelNames := funk.Keys(channelValues[0]).([]string)

				for _, expChanName := range channelNames {
					if !funk.ContainsString(providedChannelNames, expChanName.(string)) {
						invalid = true
						break;
					}
				}
			}

			if invalid {
				out.Map("valid").Push(false)
				out.Map("channels").PushBOS()
				out.Map("channels").PushEOS()
				return;
			}

			outStream := out.Map("channels").Stream()
			out.Map("valid").Push(true)
			//out.Map("channels").Stream().Push(channelValues)
			out.Map("channels").PushBOS()
			for _, kv := range channelValues {
				for k, v := range kv {
					outStream.Map(k).Push(v)
				}
			}
			out.Map("channels").PushEOS()

		}
	},
}
