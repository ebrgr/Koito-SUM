package imagepipeline

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"sync"
	"testing"
	"time"

	webp "github.com/SeriousBug/webp-go-pure/std"
)

// fixture encodes an asymmetric translucent image as PNG.
func fixture(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, color.NRGBA{uint8(x * 7), uint8(y * 9), 120, uint8(50 + x%200)})
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

// TestGeometryAndAlpha checks small, thin and normal inputs through the codec.
func TestGeometryAndAlpha(t *testing.T) {
	for _, tc := range []struct{ w, h, size, outw, outh int }{{80, 48, 300, 80, 48}, {100, 500, 300, 300, 300}, {80, 48, 0, 48, 48}, {40, 40, 16, 16, 16}} {
		out, err := Transform(context.Background(), bytes.NewReader(fixture(t, tc.w, tc.h)), tc.size)
		if err != nil {
			t.Fatal(err)
		}
		img, err := webp.DecodeBytes(out)
		if err != nil {
			t.Fatal(err)
		}
		if img.Bounds().Dx() != tc.outw || img.Bounds().Dy() != tc.outh {
			t.Fatalf("%+v: got %v", tc, img.Bounds())
		}
		_, _, _, a := img.At(0, 0).RGBA()
		if a == 0xffff {
			t.Fatalf("alpha lost: %+v", tc)
		}
	}
}

// TestInputPolicy checks admission, corrupt input and dimension boundaries.
func TestInputPolicy(t *testing.T) {
	for _, d := range []struct {
		w, h  int
		valid bool
	}{{1, 1, true}, {10000, 4000, true}, {8000, 5000, true}, {10001, 1, false}, {8000, 5001, false}, {0, 1, false}, {-1, 1, false}} {
		if (validateDimensions(d.w, d.h) == nil) != d.valid {
			t.Fatalf("%+v", d)
		}
	}
	// A valid PNG signature and IHDR with an oversized width is rejected before decode.
	data := fixture(t, 2, 2)
	binary.BigEndian.PutUint32(data[16:20], 10001)
	binary.BigEndian.PutUint32(data[29:33], crc32.ChecksumIEEE(data[12:29]))
	if _, err := Transform(context.Background(), bytes.NewReader(data), 0); !errors.Is(err, ErrDimensions) {
		t.Fatalf("oversized header: %v", err)
	}
	if _, err := Transform(context.Background(), io.LimitReader(zeroReader{}, MaxBytes+1), 0); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("bytes: %v", err)
	}
	if _, err := Transform(context.Background(), bytes.NewReader([]byte("GIF89a\x01\x00\x01\x00\x00\x00\x00")), 0); !errors.Is(err, ErrFormat) {
		t.Fatalf("GIF: %v", err)
	}
	if _, err := Transform(context.Background(), bytes.NewReader(data[:20]), 0); err == nil {
		t.Fatal("truncated image accepted")
	}
	if _, err := Transform(context.Background(), bytes.NewReader(data), 1201); !errors.Is(err, ErrSize) {
		t.Fatalf("size: %v", err)
	}
}

type zeroReader struct{}

// Read provides a stream without allocating the tested byte limit.
func (zeroReader) Read(p []byte) (int, error) { clear(p); return len(p), nil }

// TestAnimatedWebPRejectsBeforeDecode constructs a VP8X animation header.
func TestAnimatedWebPRejectsBeforeDecode(t *testing.T) {
	data := []byte{'R', 'I', 'F', 'F', 22, 0, 0, 0, 'W', 'E', 'B', 'P', 'V', 'P', '8', 'X', 10, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	if _, err := Transform(context.Background(), bytes.NewReader(data), 0); !errors.Is(err, ErrAnimation) {
		t.Fatalf("animation: %v", err)
	}
}

type gatedReader struct {
	entered chan struct{}
	release chan struct{}
	data    *bytes.Reader
	once    sync.Once
}

// Read holds the active slot so that waiting, overload, and cancellation can be observed.
func (r *gatedReader) Read(p []byte) (int, error) {
	r.once.Do(func() { close(r.entered); <-r.release })
	return r.data.Read(p)
}

// TestAPNGRejectsAnimation verifies that animated PNG cannot bypass the allowlist.
func TestAPNGRejectsAnimation(t *testing.T) {
	data := fixture(t, 2, 2)
	chunk := []byte{0, 0, 0, 8, 'a', 'c', 'T', 'L', 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0}
	binary.BigEndian.PutUint32(chunk[16:20], crc32.ChecksumIEEE(chunk[4:16]))
	data = append(append(append([]byte{}, data[:33]...), chunk...), data[33:]...)
	if _, err := Transform(context.Background(), bytes.NewReader(data), 0); !errors.Is(err, ErrAnimation) {
		t.Fatalf("APNG: %v", err)
	}
}

// TestAdmission demonstrates one active reader, four waiters and cancellable waiting.
func TestAdmission(t *testing.T) {
	reader := &gatedReader{make(chan struct{}), make(chan struct{}), bytes.NewReader(fixture(t, 4, 4)), sync.Once{}}
	done := make(chan error, 5)
	go func() { _, err := Transform(context.Background(), reader, 0); done <- err }()
	<-reader.entered
	var cancels []context.CancelFunc
	for range 4 {
		ctx, cancel := context.WithCancel(context.Background())
		cancels = append(cancels, cancel)
		go func() { _, err := Transform(ctx, bytes.NewReader(nil), 0); done <- err }()
	}
	deadline := time.After(5 * time.Second)
	for len(slots) != 5 {
		select {
		case <-deadline:
			t.Fatal("waiters failed to enter")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	if _, err := Transform(context.Background(), bytes.NewReader(nil), 0); !errors.Is(err, ErrBusy) {
		t.Fatalf("overload: %v", err)
	}
	for _, cancel := range cancels {
		cancel()
	}
	for range 4 {
		if err := <-done; !errors.Is(err, context.Canceled) {
			t.Fatalf("waiter: %v", err)
		}
	}
	close(reader.release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

// TestJPEGOrientations checks all EXIF mappings through the public pipeline.
func TestJPEGOrientations(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 40, 24))
	for y := 0; y < 24; y++ {
		for x := 0; x < 40; x++ {
			img.SetNRGBA(x, y, color.NRGBA{uint8(x * 6), uint8(y * 10), 0, 255})
		}
	}
	var b bytes.Buffer
	if err := jpeg.Encode(&b, img, &jpeg.Options{Quality: 100}); err != nil {
		t.Fatal(err)
	}
	for orientation := 1; orientation <= 8; orientation++ {
		app := []byte{'E', 'x', 'i', 'f', 0, 0, 'I', 'I', 42, 0, 8, 0, 0, 0, 1, 0, 18, 1, 3, 0, 1, 0, 0, 0, byte(orientation), 0, 0, 0, 0, 0, 0, 0}
		data := append([]byte{}, b.Bytes()[:2]...)
		data = append(data, 255, 225, 0, byte(len(app)+2))
		data = append(data, app...)
		data = append(data, b.Bytes()[2:]...)
		// Target larger than both axes preserves the oriented rectangle.
		out, err := Transform(context.Background(), bytes.NewReader(data), 100)
		if err != nil {
			t.Fatalf("orientation %d: %v", orientation, err)
		}
		got, err := webp.DecodeBytes(out)
		if err != nil {
			t.Fatal(err)
		}
		if orientation >= 5 && got.Bounds().Size() != image.Pt(24, 40) {
			t.Fatalf("orientation %d bounds %v", orientation, got.Bounds())
		}
		if orientation < 5 && got.Bounds().Size() != image.Pt(40, 24) {
			t.Fatalf("orientation %d bounds %v", orientation, got.Bounds())
		}
		// Sample a corner whose channel values encode original coordinates.
		dx, dy := 2, 2
		switch orientation {
		case 2:
			dx = 37
		case 3:
			dx, dy = 37, 21
		case 4:
			dy = 21
		case 5:
			dx, dy = 2, 2
		case 6:
			dx, dy = 21, 2
		case 7:
			dx, dy = 21, 37
		case 8:
			dx, dy = 2, 37
		}
		r, g, _, _ := got.At(dx, dy).RGBA()
		if r>>8 > 35 || g>>8 > 45 {
			t.Fatalf("orientation %d corner: %d %d", orientation, r>>8, g>>8)
		}
	}
}

// TestWithDecodedSharesAdmission035 ensures the decode callback holds the global slot.
func TestWithDecodedSharesAdmission035(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	data := fixture(t, 4, 4)
	go func() {
		done <- WithDecoded(context.Background(), bytes.NewReader(data), func(image.Image) error { close(entered); <-release; return nil })
	}()
	<-entered
	ctx, cancel := context.WithCancel(context.Background())
	waiting := make(chan error, 1)
	go func() { _, err := Transform(ctx, bytes.NewReader(data), 0); waiting <- err }()
	deadline := time.After(3 * time.Second)
	for len(slots) != 2 {
		select {
		case <-deadline:
			close(release)
			t.Fatal("transform did not wait")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	cancel()
	if err := <-waiting; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

// TestWithDecodedCallbackError035 preserves callback failures and releases capacity.
func TestWithDecodedCallbackError035(t *testing.T) {
	input := fixture(t, 4, 4)
	sentinel := errors.New("callback failed")
	err := WithDecoded(context.Background(), bytes.NewReader(input), func(image.Image) error { return sentinel })
	if !errors.Is(err, sentinel) {
		t.Fatalf("callback error: %v", err)
	}
	if _, err := Transform(context.Background(), bytes.NewReader(input), 0); err != nil {
		t.Fatalf("capacity not released: %v", err)
	}
}
