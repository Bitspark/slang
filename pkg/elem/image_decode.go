package elem

import (
	"bytes"
	"image"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var imageDecodeCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: uuid.MustParse("4b082c52-9a99-472f-9277-f5ca9651dbfb"),
		Meta: core.BlueprintMetaDef{
			Name:             "decode image",
			ShortDescription: "reads an encoded image binary and emits its pixels as stream of rgb values",
			Icon:             "file-image",
			Tags:             []string{"file"},
			DocURL:           "https://bitspark.de/slang/docs/operator/decode-image",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "binary",
				},
				Out: core.TypeDef{
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
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			ib := i.(core.Binary)
			img, format, err := image.Decode(bytes.NewReader(ib))
			if err != nil {
				out.Push(nil)
				continue
			}

			out.Map("format").Push(format)

			width := img.Bounds().Max.X - img.Bounds().Min.X
			height := img.Bounds().Max.Y - img.Bounds().Min.Y

			out.Map("width").Push(float64(width))
			out.Map("height").Push(float64(height))

			out.Map("pixels").PushBOS()
			pixels := out.Map("pixels").Stream()
			for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
				for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
					red, green, blue, alpha := img.At(x, y).RGBA()
					pixels.Push(map[string]interface{}{
						"red":   float64(red),
						"green": float64(green),
						"blue":  float64(blue),
						"alpha": float64(alpha),
					})
				}
			}
			out.Map("pixels").PushEOS()
		}
	},
}
