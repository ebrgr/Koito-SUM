// This executable reproduces the production bimg options without a database.
package main

import (
	"github.com/h2non/bimg"
	"koito-images031/worker"
)

func inspect(data []byte) (int, int, error) { s, e := bimg.Size(data); return s.Width, s.Height, e }
func transform(data []byte, size, effort int) ([]byte, error) {
	px, q := size, 85
	if size == 0 {
		s, e := bimg.Size(data)
		if e != nil {
			return nil, e
		}
		px = min(s.Width, s.Height, 1200)
		q = 100
	}
	return bimg.NewImage(data).Process(bimg.Options{Width: px, Height: px, Crop: true, Quality: q, StripMetadata: true, Type: bimg.WEBP})
}
func main() { worker.Run(inspect, transform) }
