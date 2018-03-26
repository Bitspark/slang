package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/tealeg/xlsx"
)

var xlsxReadOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		Services: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.PortDef{
					Type: "string",
				},
				Out: core.PortDef{
					Type: "stream",
					Stream: &core.PortDef{
						Type: "map",
						Map: map[string]*core.PortDef{
							"name": {
								Type: "string",
							},
							"content": {
								Type: "stream",
								Stream: &core.PortDef{
									Type: "stream",
									Stream: &core.PortDef{
										Type: "string",
									},
								},
							},
						},
					},
				},
			},
		},
		Delegates: map[string]*core.DelegateDef{},
	},
	oFunc: func(srvs map[string]*core.Service, dels map[string]*core.Delegate, store interface{}) {
		in := srvs[core.MAIN_SERVICE].In()
		out := srvs[core.MAIN_SERVICE].Out()
		for {
			filename, i := in.PullString()
			if i != nil {
				out.Push(i)
				continue
			}
			xlsxFile, err := xlsx.OpenFile(filename)
			if err != nil {
				panic(err)
			}
			out.PushBOS()
			for _, sheet := range xlsxFile.Sheets {
				out.Stream().Map("name").Push(sheet.Name)
				out.Stream().Map("content").PushBOS()
				for _, row := range sheet.Rows {
					out.Stream().Map("content").Stream().PushBOS()
					for _, col := range row.Cells {
						out.Stream().Map("content").Stream().Stream().Push(col.Value)
					}
					out.Stream().Map("content").Stream().PushEOS()
				}
				out.Stream().Map("content").PushEOS()
			}
			out.PushEOS()
		}
	},
}
