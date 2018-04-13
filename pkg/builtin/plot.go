package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/plotutil"
	"strconv"
	"fmt"
)

var plotOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"graph_{ids}": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type: "map",
								Map: map[string]*core.TypeDef{
									"x": {
										Type: "number",
									},
									"y": {
										Type: "number",
									},
								},
							},
						},
						"filename": {
							Type: "string",
						},
					},
				},
				Out: core.TypeDef{
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{},
		PropertyDefs: map[string]*core.TypeDef{
			"ids": {
				Type: "stream",
				Stream: &core.TypeDef{
					Type: "string",
				},
			},
			"graphs": {
				Type: "stream",
				Stream: &core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"label": {
							Type: "string",
						},
						"color": {
							Type: "string",
						},
					},
				},
			},
			"scaleX": {
				Type: "string",
			},
			"scaleY": {
				Type: "string",
			},
			"minX": {
				Type: "number",
			},
			"maxX": {
				Type: "number",
			},
			"minY": {
				Type: "number",
			},
			"maxY": {
				Type: "number",
			},
		},
	},
	oFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		idsI := op.Property("ids").([]interface{})
		var ids []string
		for _, idI := range idsI {
			ids = append(ids, idI.(string))
		}
		graphsI := op.Property("graphs").([]interface{})
		var graphs []map[string]string
		for _, graphI := range graphsI {
			graphMap := graphI.(map[string]interface{})
			entry := make(map[string]string)
			for k, v := range graphMap {
				entry[k] = v.(string)
			}
			graphs = append(graphs, entry)
		}
		scaleX := op.Property("scaleX").(string)
		scaleY := op.Property("scaleY").(string)
		minX := op.Property("minX").(float64)
		maxX := op.Property("maxX").(float64)
		minY := op.Property("minX").(float64)
		maxY := op.Property("maxX").(float64)
		for {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
			}

			data := i.(map[string]interface{})

			p, err := plot.New()
			if err != nil {
				panic(err)
			}

			if scaleX == "linear" {
				p.X.Scale = plot.LinearScale{}
			} else if scaleX == "logarithmic" {
				p.X.Scale = plot.LogScale{}
			}

			if scaleY == "linear" {
				p.Y.Scale = plot.LinearScale{}
			} else if scaleY == "logarithmic" {
				p.Y.Scale = plot.LogScale{}
			}

			p.X.Min = minX
			p.X.Max = maxX
			p.Y.Min = minY
			p.Y.Max = maxY

			for idx, id := range ids {
				graphPoints := data["graph_"+id].([]interface{})
				pts := make(plotter.XYs, len(graphPoints))
				for jdx, graphPoint := range graphPoints {
					points := graphPoint.(map[string]interface{})
					pts[jdx].X, _ = strconv.ParseFloat(fmt.Sprintf("%v", points["x"]), 64)
					pts[jdx].Y, _ = strconv.ParseFloat(fmt.Sprintf("%v", points["y"]), 64)
				}

				err = plotutil.AddLinePoints(p, graphs[idx]["label"], pts)
				if err != nil {
					panic(err)
				}
			}

			p.Title.Text = "Slang plot"
			p.X.Label.Text = "time"

			if err := p.Save(4*vg.Inch, 4*vg.Inch, data["filename"].(string)); err != nil {
				panic(err)
			}
		}
	},
}
