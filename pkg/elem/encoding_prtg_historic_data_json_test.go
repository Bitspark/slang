package elem

import (
	"testing"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/stretchr/testify/require"
)

func Test_PRTGHistData__IsRegistered(t *testing.T) {
	Init()
	a := assertions.New(t)

	ocFork := getBuiltinCfg(encodingPRTGHistDataId)
	a.NotNil(ocFork)
}

func Test_PRTGHistData__Best_Case(t *testing.T) {
	Init()
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: encodingPRTGHistDataId,
			Properties: core.Properties{
				"ChannelNames": []interface{}{"Power from PV", "Solar Energy Today"},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()


	jsonData := `{
		"prtg-version": "23.1.82.2175",
		"treesize": 1200,
		"histdata": [
		  {
			"datetime": "27.03.2023 00:00:00",
			" Power from PV ": "",
			" Solar Energy Today ": "",
			"coverage": "0 %"
		  },
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
	}`


	o.Main().In().Push(core.Binary(jsonData))

	a.PortPushes(true, o.Main().Out().Map("valid"))
	a.PortPushes([]interface{}{
		map[string]interface{}{
			"Power from PV":      22.691,
			"Solar Energy Today": 112.5,
		},
		map[string]interface{}{
			"Power from PV":      21.4,
			"Solar Energy Today": 112.9,
		},
		map[string]interface{}{
			"Power from PV":      21.863,
			"Solar Energy Today": 113.2,
		},
	}, o.Main().Out().Map("channels"))
}

func Test_PRTGHistData__Channel_is_subset_of_possible_fields(t *testing.T) {
	Init()
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: encodingPRTGHistDataId,
			Properties: core.Properties{
				// There is no PRTG Channel called XYZ123
				"ChannelNames": []interface{}{"Power from PV"},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()


	jsonData := `{
		"prtg-version": "23.1.82.2175",
		"treesize": 1200,
		"histdata": [
		  {
			"datetime": "27.03.2023 00:00:00",
			" Power from PV ": "",
			" Solar Energy Today ": "",
			"coverage": "0 %"
		  },
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
	}`

	o.Main().In().Push(core.Binary(jsonData))

	a.PortPushes(true, o.Main().Out().Map("valid"))
	a.PortPushes([]interface{}{
		map[string]interface{}{
			"Power from PV":      22.691,
		},
		map[string]interface{}{
			"Power from PV":      21.4,
		},
		map[string]interface{}{
			"Power from PV":      21.863,
		},
	}, o.Main().Out().Map("channels"))
}

func Test_PRTGHistData__Expect_more_channels_than_existing_fields(t *testing.T) {
	Init()
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: encodingPRTGHistDataId,
			Properties: core.Properties{
				// There is no PRTG Channel called XYZ123
				"ChannelNames": []interface{}{"Power from PV", "XXX"},
			},
		},
	)
	require.NoError(t, err)

	o.Main().Out().Bufferize()
	o.Start()


	jsonData := `{
		"prtg-version": "23.1.82.2175",
		"treesize": 1200,
		"histdata": [
		  {
			"datetime": "27.03.2023 00:00:00",
			" Power from PV ": "",
			"coverage": "0 %"
		  },
		  {
			"datetime": "27.03.2023 12:37:00",
			" Power from PV ": 22.6910,
			"coverage": "100 %"
		  },
		  {
			"datetime": "27.03.2023 12:38:00",
			" Power from PV ": 21.4000,
			"coverage": "100 %"
		  },
		  {
			"datetime": "27.03.2023 12:39:00",
			" Power from PV ": 21.8630,
			"coverage": "100 %"
		  }
		]
	}`

	o.Main().In().Push(core.Binary(jsonData))

	a.PortPushes(false, o.Main().Out().Map("valid"))
	a.PortPushes([]interface{}{}, o.Main().Out().Map("channels"))
}