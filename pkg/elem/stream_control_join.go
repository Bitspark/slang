package elem

import (
	"fmt"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var streamCtrlJoinCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: uuid.MustParse("fb174c53-80bd-4e29-955a-aafe33ebfb31"),
		Meta: core.BlueprintMetaDef{
			Name:             "join",
			ShortDescription: "joins two streams",
			Icon:             "layer-plus",
			Tags:             []string{"stream"},
			DocURL:           "https://bitspark.de/slang/docs/operator/concatenate",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"stream_{streams}": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type:    "generic",
								Generic: "itemType",
							},
						},
					},
				},
				Out: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type:    "generic",
						Generic: "itemType",
					},
				},
			},
		},
		PropertyDefs: core.PropertyMap{
			"streams": {
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
		indexesProp := op.Property("streams").([]interface{})
		streams := make([]*core.Port, len(indexesProp))
		for i, idxProp := range indexesProp {
			streams[i] = in.Map("stream_" + idxProp.(string))
		}
		for !op.CheckStop() {
			item := streams[0].Stream().Pull()
			// Forward "other" BOS arriving at all Ports
			if !streams[0].OwnBOS(item) {
				for i := 1; i < len(streams); i++ {
					streams[i].Stream().Pull()
				}
				out.Push(item)
				continue
			}


			// Pull BOS of all stream ports
			for i := 1; i < len(streams); i++ {
				streams[i].PullBOS()
			}

			out.PushBOS()
			// Pull items from each stream one by one
			// skip pulling item from a stream port when EOS has arrived
			allDone := false
			streamDone := []bool{}
			for _ = range streams {
				streamDone = append(streamDone, false)
			}

			for !allDone {
				// I assume all streams are done,
				// if not inside for next for it will be set false
				allDone = true

				for i, s := range streams {
					if streamDone[i] {
						continue
					}

					allDone = false

					if item, ok := s.Stream().Poll(); ok {
						if s.OwnEOS(item) {
							streamDone[i] = true
							continue
						}
						out.Stream().Push(item)
					}
				}
			}


			/*
			for i := 0; i < len(streams); i++ {
				for {
					item = streams[i].Stream().Pull()
					if streams[i].OwnEOS(item) {
						if i+1 < len(streams) {
							streams[i+1].PullBOS()
						}
						break
					}
					out.Stream().Push(item)
				}
			}
			*/


			fmt.Println("EOS")
			out.PushEOS()
		}
	},
}
