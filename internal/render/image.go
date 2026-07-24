package render

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
)

// ImageMetadata holds detected dimensions and aspect ratio to prevent CLS.
type ImageMetadata struct {
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	Format      string  `json:"format"`
	AspectRatio float64 `json:"aspectRatio"`
	AspectCSS   string  `json:"aspectCss"`
}

// DetectImageMetadata reads image headers without loading the full pixel array into memory.
func DetectImageMetadata(filePath string) (ImageMetadata, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return ImageMetadata{}, fmt.Errorf("zyra/render/image: failed to open image %s: %w", filePath, err)
	}
	defer f.Close()

	return DetectImageMetadataFromReader(f)
}

// DetectImageMetadataFromReader decodes image headers from a reader.
func DetectImageMetadataFromReader(r io.Reader) (ImageMetadata, error) {
	cfg, format, err := image.DecodeConfig(r)
	if err != nil {
		return ImageMetadata{}, fmt.Errorf("zyra/render/image: failed to decode image header: %w", err)
	}

	aspectRatio := 1.0
	aspectCSS := "1 / 1"
	if cfg.Height > 0 {
		aspectRatio = float64(cfg.Width) / float64(cfg.Height)
		aspectCSS = fmt.Sprintf("%d / %d", cfg.Width, cfg.Height)
	}

	return ImageMetadata{
		Width:       cfg.Width,
		Height:      cfg.Height,
		Format:      format,
		AspectRatio: aspectRatio,
		AspectCSS:   aspectCSS,
	}, nil
}
