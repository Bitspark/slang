package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"github.com/Bitspark/go-funk"
	"image/png"
	"bytes"
	"bufio"
	"github.com/Bitspark/slang/pkg/utils"
	"image/jpeg"
)

var imageEncodeCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"height": {
							Type: "number",
						},
						"width": {
							Type: "number",
						},
						"format": {
							Type: "string",
						},
						"pixels": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type: "map",
								Map: map[string]*core.TypeDef{
									"red": {
										Type: "number",
									},
									"green": {
										Type: "number",
									},
									"blue": {
										Type: "number",
									},
									"alpha": {
										Type: "number",
									},
								},
							},
						},
					},
				},
				Out: core.TypeDef{
					Type: "binary",
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		formats := []string{"jpeg", "png"}
		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			if i == nil {
				out.Push(nil)
				continue
			}

			im := i.(map[string]interface{})
			width := im["width"].(float64)
			height := im["height"].(float64)
			format := im["format"].(string)

			if !funk.Contains(formats, format) {
				out.Push(nil)
				continue
			}

			img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
			pixels := im["pixels"].([]interface{})
			for y := 0; y < int(height); y++ {
				for x := 0; x < int(width); x++ {
					idx := y * int(width) + x
					pixel := pixels[idx].(map[string]interface{})
					img.Set(x, y,
						color.RGBA64{
							R: uint16(pixel["red"].(float64)),
							G: uint16(pixel["green"].(float64)),
							B: uint16(pixel["blue"].(float64)),
							A: uint16(pixel["alpha"].(float64)),
						},
					)
				}
			}

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)

			switch format {
			case "png":
				png.Encode(writer, img)
				break
			case "jpeg":
				jpeg.Encode(writer, img, &jpeg.Options{
					Quality: 100,
				})
				break
			}

			out.Push(utils.Binary(b.Bytes()))
		}
	},
}
