package render

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestDetectImageMetadata(t *testing.T) {
	// Create a test 200x100 PNG image in memory
	img := image.NewRGBA(image.Rect(0, 0, 200, 100))
	img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode test image: %v", err)
	}

	meta, err := DetectImageMetadataFromReader(&buf)
	if err != nil {
		t.Fatalf("DetectImageMetadataFromReader failed: %v", err)
	}

	if meta.Width != 200 {
		t.Errorf("expected width 200, got %d", meta.Width)
	}
	if meta.Height != 100 {
		t.Errorf("expected height 100, got %d", meta.Height)
	}
	if meta.Format != "png" {
		t.Errorf("expected format png, got %s", meta.Format)
	}
	if meta.AspectCSS != "200 / 100" {
		t.Errorf("expected AspectCSS '200 / 100', got %s", meta.AspectCSS)
	}
	if meta.AspectRatio != 2.0 {
		t.Errorf("expected AspectRatio 2.0, got %f", meta.AspectRatio)
	}
}
