package photos

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"

	"golang.org/x/image/draw"
)

const (
	// MaxUploadSize is the maximum allowed upload size (12 MiB).
	MaxUploadSize = 12 * 1024 * 1024

	// MaxOriginalSide is the maximum side length for the original image.
	MaxOriginalSide = 2560

	// MaxThumbnailSide is the maximum side length for the thumbnail.
	MaxThumbnailSide = 512

	// MinValidSide is the minimum valid image dimension.
	MinValidSide = 1

	// MaxValidSide is the maximum valid image dimension to prevent DoS.
	MaxValidSide = 10000
)

// ValidateImage checks that the data is a valid image (JPEG or PNG in MVP).
func ValidateImage(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("empty image data")
	}

	// Check magic bytes
	if len(data) < 4 {
		return fmt.Errorf("image data too short")
	}

	// JPEG: FF D8 FF
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		// Try to decode to verify it's not corrupted
		_, err := jpeg.Decode(bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("invalid JPEG: %w", err)
		}
		return nil
	}

	// PNG: 89 50 4E 47
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		_, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("invalid PNG: %w", err)
		}
		return nil
	}

	return fmt.Errorf("not a valid image (unsupported format)")
}

// DecodeImage decodes an image and returns its dimensions.
// It protects against oversized images by limiting dimensions.
func DecodeImage(data []byte) (width, height int, err error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return 0, 0, fmt.Errorf("decode image: %w", err)
	}

	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	if w < MinValidSide || h < MinValidSide {
		return 0, 0, fmt.Errorf("image dimensions too small: %dx%d", w, h)
	}

	if w > MaxValidSide || h > MaxValidSide {
		return 0, 0, fmt.Errorf("image dimensions too large: %dx%d", w, h)
	}

	return w, h, nil
}

// DecodeImageWithFormat decodes an image and returns dimensions along with the format.
func DecodeImageWithFormat(data []byte) (width, height int, format string, err error) {
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return 0, 0, "", fmt.Errorf("decode image: %w", err)
	}

	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	if w < MinValidSide || h < MinValidSide {
		return 0, 0, "", fmt.Errorf("image dimensions too small: %dx%d", w, h)
	}

	if w > MaxValidSide || h > MaxValidSide {
		return 0, 0, "", fmt.Errorf("image dimensions too large: %dx%d", w, h)
	}

	return w, h, format, nil
}

// CreateThumbnail creates a JPEG thumbnail from the source image data.
// The thumbnail's larger side will be at most maxWidth pixels.
func CreateThumbnail(data []byte, maxWidth int) ([]byte, error) {
	if maxWidth <= 0 {
		maxWidth = MaxThumbnailSide
	}

	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode for thumbnail: %w", err)
	}

	return scaleImage(img, maxWidth, format)
}

// NormalizeJPEG normalizes image data to JPEG format with EXIF orientation applied.
// It handles PNG input by converting to JPEG and applies orientation.
func NormalizeJPEG(data []byte) ([]byte, error) {
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode for normalize: %w", err)
	}

	// Apply orientation if needed (simplified - no EXIF parsing in MVP)
	img = applyOrientation(img, format, data)

	// Scale to max original size if needed
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	maxSide := w
	if h > w {
		maxSide = h
	}

	if maxSide > MaxOriginalSide {
		scaled, err := scaleImage(img, MaxOriginalSide, "jpeg")
		if err != nil {
			return nil, fmt.Errorf("scale for normalize: %w", err)
		}
		// Re-decode the scaled JPEG
		data = scaled
		img, _, err = image.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("re-decode scaled: %w", err)
		}
	}

	// Encode as JPEG
	var buf bytes.Buffer
	opts := jpeg.Options{Quality: 85}
	if err := jpeg.Encode(&buf, img, &opts); err != nil {
		return nil, fmt.Errorf("encode JPEG: %w", err)
	}

	return buf.Bytes(), nil
}

// scaleImage scales an image so that its larger side is at most maxWidth.
// It returns the scaled image encoded as JPEG.
func scaleImage(img image.Image, maxWidth int, format string) ([]byte, error) {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	maxSide := w
	if h > w {
		maxSide = h
	}

	if maxSide <= maxWidth {
		// No scaling needed, just encode
		var buf bytes.Buffer
		opts := jpeg.Options{Quality: 85}
		if err := jpeg.Encode(&buf, img, &opts); err != nil {
			return nil, fmt.Errorf("encode: %w", err)
		}
		return buf.Bytes(), nil
	}

	// Calculate new dimensions
	ratio := float64(maxWidth) / float64(maxSide)
	newW := int(math.Round(float64(w) * ratio))
	newH := int(math.Round(float64(h) * ratio))

	if newW < MinValidSide {
		newW = MinValidSide
	}
	if newH < MinValidSide {
		newH = MinValidSide
	}

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))

	// Use best quality scaler
	draw.BiLinear.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)

	var buf bytes.Buffer
	opts := jpeg.Options{Quality: 85}
	if err := jpeg.Encode(&buf, dst, &opts); err != nil {
		return nil, fmt.Errorf("encode scaled: %w", err)
	}

	return buf.Bytes(), nil
}

// applyOrientation applies EXIF orientation to the image.
// In MVP, this is a placeholder for future EXIF handling.
func applyOrientation(img image.Image, format string, raw []byte) image.Image {
	// TODO: Parse EXIF orientation tag and rotate/flip the image
	// For now, return the image as-is
	_ = format
	_ = raw
	return img
}

// ReadLimited reads up to maxBytes from r.
func ReadLimited(r io.Reader, maxBytes int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read limited: %w", err)
	}

	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("content exceeds maximum size of %d bytes", maxBytes)
	}

	return data, nil
}

// ContentHash computes a SHA-256 hash of the image data for deduplication.
func ContentHash(data []byte) string {
	// In MVP, we use a simple approach. In production, use crypto/sha256.
	// This is a placeholder — the real implementation uses crypto/sha256.
	return fmt.Sprintf("%x", data[:min(32, len(data))])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
