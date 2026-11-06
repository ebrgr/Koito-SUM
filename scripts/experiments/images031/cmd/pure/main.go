// This executable studies the candidate without importing bimg or the old WebP decoder.
package main

import (
	"bytes"
	"fmt"
	codec "github.com/SeriousBug/webp-go-pure"
	webp "github.com/SeriousBug/webp-go-pure/std"
	"github.com/disintegration/imaging"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"koito-images031/worker"
)

// inspect enforces an explicit product allowlist before image.Decode can select a codec.
func inspect(data []byte) (int, int, error) {
	if len(data) >= 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
		f, e := codec.Features(data)
		if e != nil {
			return 0, 0, e
		}
		if f.HasAnimation {
			return f.Width, f.Height, fmt.Errorf("animation-rejected")
		}
		return f.Width, f.Height, nil
	}
	c, f, e := image.DecodeConfig(bytes.NewReader(data))
	if e != nil {
		return 0, 0, e
	}
	if f != "jpeg" && f != "png" {
		return c.Width, c.Height, fmt.Errorf("format-rejected:%s", f)
	}
	return c.Width, c.Height, nil
}

// transform reproduces the observed no-enlarge branch instead of blindly filling a square.
func transform(data []byte, size, effort int) ([]byte, error) {
	c, _, e := image.DecodeConfig(bytes.NewReader(data))
	if e != nil {
		return nil, e
	}
	px := size
	q := 85
	if size == 0 {
		px = min(c.Width, c.Height, 1200)
		q = 100
	}
	img, e := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
	if e != nil {
		return nil, e
	}
	b := img.Bounds()
	var dst image.Image
	if b.Dx() < px && b.Dy() < px {
		dst = img
	} else {
		dst = imaging.Fill(img, px, px, imaging.Center, imaging.Lanczos)
	}
	return webp.EncodeBytes(dst, &webp.Options{Quality: q, Effort: effort})
}
func main() { worker.Run(inspect, transform) }
