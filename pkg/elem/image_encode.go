package elem

import (
	"bufio"
	"bytes"
	"github.com/Bitspark/go-funk"
	"github.com/Bitspark/slang/pkg/core"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
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
			i := in.Map("height").Pull()
			if core.IsMarker(i) || i == nil {
				in.Map("width").Pull()
				in.Map("format").Pull()
				in.Map("pixels").Pull()
				out.Push(i)
				continue
			}

			height := i.(float64)
			width := in.Map("width").Pull().(float64)
			format := in.Map("format").Pull().(string)

			if !funk.Contains(formats, format) {
				in.Map("pixels").Pull()
				out.Push(nil)
				continue
			}

			in.Map("pixels").PullBOS()
			pixelStream := in.Map("pixels").Stream()
			img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
			for y := 0; y < int(height); y++ {
				for x := 0; x < int(width); x++ {
					pixel := pixelStream.Pull().(map[string]interface{})
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
			in.Map("pixels").PullEOS()

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

			out.Push(core.Binary(b.Bytes()))
		}
	},
}
