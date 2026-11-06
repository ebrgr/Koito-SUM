package main

import (
	"bytes"
	webp "github.com/SeriousBug/webp-go-pure/std"
	"github.com/disintegration/imaging"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// TestNonZeroBounds compares a subimage with an equivalent normalized source.
func TestNonZeroBounds(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 40, 40))
	for y := 0; y < 40; y++ {
		for x := 0; x < 40; x++ {
			src.SetNRGBA(x, y, color.NRGBA{uint8(x * 5), uint8(y * 5), 80, uint8(50 + x + y)})
		}
	}
	sub := src.SubImage(image.Rect(7, 9, 32, 28))
	a := imaging.Fill(sub, 16, 16, imaging.Center, imaging.Lanczos)
	b := imaging.Fill(imaging.Clone(sub), 16, 16, imaging.Center, imaging.Lanczos)
	if !bytes.Equal(a.Pix, b.Pix) {
		t.Fatal("bounds changed result")
	}
}

// TestAlphaRoundTrip verifies the encoder keeps alpha after the chosen transform.
func TestAlphaRoundTrip(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 40, 30))
	for y := 0; y < 30; y++ {
		for x := 0; x < 40; x++ {
			src.SetNRGBA(x, y, color.NRGBA{250, 60, 20, uint8(x * 6)})
		}
	}
	transformed := imaging.Fill(src, 16, 16, imaging.Center, imaging.Lanczos)
	data, e := webp.EncodeBytes(transformed, &webp.Options{Quality: 85, Effort: -1})
	if e != nil {
		t.Fatal(e)
	}
	if directory := os.Getenv("ALPHA_EVIDENCE_DIR"); directory != "" {
		f, err := os.Create(filepath.Join(directory, "alpha-reference.png"))
		if err != nil {
			t.Fatal(err)
		}
		if err = png.Encode(f, transformed); err != nil {
			t.Fatal(err)
		}
		if err = f.Close(); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(directory, "alpha-output.webp"), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	decoded, e := webp.DecodeBytes(data)
	if e != nil {
		t.Fatal(e)
	}
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			if color.NRGBAModel.Convert(decoded.At(x, y)).(color.NRGBA).A != transformed.NRGBAAt(x, y).A {
				t.Fatal("alpha changed")
			}
		}
	}
}

// TestJPEGOrientations checks all EXIF mappings against literal coordinate rules.
func TestJPEGOrientations(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 20, 12))
	for y := 0; y < 12; y++ {
		for x := 0; x < 20; x++ {
			src.SetNRGBA(x, y, color.NRGBA{uint8(x * 12), uint8(y * 20), uint8((x + y) * 7), 255})
		}
	}
	var encoded bytes.Buffer
	if e := jpeg.Encode(&encoded, src, &jpeg.Options{Quality: 100}); e != nil {
		t.Fatal(e)
	}
	original, e := jpeg.Decode(bytes.NewReader(encoded.Bytes()))
	if e != nil {
		t.Fatal(e)
	}
	for orientation := 1; orientation <= 8; orientation++ {
		app := []byte{'E', 'x', 'i', 'f', 0, 0, 'I', 'I', 42, 0, 8, 0, 0, 0, 1, 0, 18, 1, 3, 0, 1, 0, 0, 0, byte(orientation), 0, 0, 0, 0, 0, 0, 0}
		data := append([]byte{}, encoded.Bytes()[:2]...)
		data = append(data, 255, 225, byte((len(app)+2)>>8), byte(len(app)+2))
		data = append(data, app...)
		data = append(data, encoded.Bytes()[2:]...)
		got, e := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
		if e != nil {
			t.Fatal(e)
		}
		want := image.Pt(20, 12)
		if orientation >= 5 {
			want = image.Pt(12, 20)
		}
		if got.Bounds().Size() != want {
			t.Fatalf("orientation%d dimensions", orientation)
		}
		for y := 0; y < 12; y++ {
			for x := 0; x < 20; x++ {
				dx, dy := x, y
				switch orientation {
				case 2:
					dx = 19 - x
				case 3:
					dx, dy = 19-x, 11-y
				case 4:
					dy = 11 - y
				case 5:
					dx, dy = y, x
				case 6:
					dx, dy = 11-y, x
				case 7:
					dx, dy = 11-y, 19-x
				case 8:
					dx, dy = y, 19-x
				}
				if color.NRGBAModel.Convert(got.At(dx, dy)) != color.NRGBAModel.Convert(original.At(x, y)) {
					t.Fatalf("orientation%d pixel%d,%d", orientation, x, y)
				}
			}
		}
	}
}
