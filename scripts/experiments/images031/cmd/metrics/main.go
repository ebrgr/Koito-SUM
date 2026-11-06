// Metrics compare opaque composites and alpha against an independent lossless reference.
package main

import (
	"bytes"
	"encoding/json"
	_ "github.com/SeriousBug/webp-go-pure/std"
	"github.com/disintegration/imaging"
	"image"
	"image/color"
	_ "image/png"
	"math"
	"os"
)

func main() {
	ref, e := imaging.Open(os.Args[1])
	if e != nil {
		panic(e)
	}
	data, e := os.ReadFile(os.Args[2])
	if e != nil {
		panic(e)
	}
	img, _, e := image.Decode(bytes.NewReader(data))
	if e != nil {
		panic(e)
	}
	b := ref.Bounds()
	if b.Size() != img.Bounds().Size() {
		json.NewEncoder(os.Stdout).Encode(map[string]any{"dimension_match": false})
		return
	}
	mse := [2]float64{}
	alphaMax := 0
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			a := color.NRGBAModel.Convert(ref.At(x+b.Min.X, y+b.Min.Y)).(color.NRGBA)
			c := color.NRGBAModel.Convert(img.At(x+img.Bounds().Min.X, y+img.Bounds().Min.Y)).(color.NRGBA)
			alphaMax = max(alphaMax, int(math.Abs(float64(a.A)-float64(c.A))))
			for k, bg := range []float64{0, 255} {
				for j, av := range []uint8{a.R, a.G, a.B} {
					cv := []uint8{c.R, c.G, c.B}[j]
					v := float64(av)*float64(a.A)/255 + bg*(1-float64(a.A)/255)
					w := float64(cv)*float64(c.A)/255 + bg*(1-float64(c.A)/255)
					mse[k] += (v - w) * (v - w)
				}
			}
		}
	}
	result := map[string]any{"dimension_match": true, "alpha_max_error": alphaMax}
	for k, name := range []string{"psnr_black_db", "psnr_white_db"} {
		v := mse[k] / float64(3*b.Dx()*b.Dy())
		if v == 0 {
			result[name] = 100.0
		} else {
			result[name] = 10 * math.Log10(255*255/v)
		}
	}
	json.NewEncoder(os.Stdout).Encode(result)
}
