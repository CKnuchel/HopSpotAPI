package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"slices"
	"strings"

	"github.com/disintegration/imaging"
)

type ImageSize struct {
	Width  int
	Height int
}

var (
	SizeOriginal  = ImageSize{Width: 1920, Height: 1080}
	SizeMedium    = ImageSize{Width: 800, Height: 600}
	SizeThumbnail = ImageSize{Width: 200, Height: 200}
)

type ProcessedImages struct {
	Original  []byte
	Medium    []byte
	Thumbnail []byte
}

func ProcessImage(reader io.Reader) (*ProcessedImages, error) {
	// Decoding the image
	img, err := imaging.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Original (max 1920x1080)
	original := resizeImage(img, SizeOriginal.Width, SizeOriginal.Height)
	originalBytes, err := encodeJPEG(original, 90)
	if err != nil {
		return nil, err
	}

	// Medium (max 800x600)
	medium := resizeImage(img, SizeMedium.Width, SizeMedium.Height)
	mediumBytes, err := encodeJPEG(medium, 85)
	if err != nil {
		return nil, err
	}

	// Thumbnail (200x200, cropped to square)
	thumbnail := imaging.Thumbnail(img, SizeThumbnail.Width, SizeThumbnail.Height, imaging.Lanczos)
	thumbnailBytes, err := encodeJPEG(thumbnail, 80)
	if err != nil {
		return nil, err
	}

	return &ProcessedImages{
		Original:  originalBytes,
		Medium:    mediumBytes,
		Thumbnail: thumbnailBytes,
	}, nil
}

// Scales the image proportionally
func resizeImage(img image.Image, maxWidth, maxHeight int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Only scale down
	if width <= maxWidth && height <= maxHeight {
		return img
	}

	// Calculate new dimensions
	ratioW := float64(maxWidth) / float64(width)
	ratioH := float64(maxHeight) / float64(height)

	ratio := ratioW
	if ratioH < ratioW {
		ratio = ratioH
	}

	newWidth := int(float64(width) * ratio)
	newHeight := int(float64(height) * ratio)

	return imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)
}

// Converts an image to JPEG bytes
func encodeJPEG(img image.Image, quality int) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := jpeg.Encode(buf, img, &jpeg.Options{Quality: quality})
	if err != nil {
		return nil, fmt.Errorf("failed to encode jpeg: %w", err)
	}

	return buf.Bytes(), nil
}

// Checks whether the MIME type is allowed
func ValidateImageType(contentType string) bool {
	allowed := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/webp",
	}

	contentType = strings.ToLower(contentType)
	return slices.Contains(allowed, contentType)
}

// GeneratePhotoPath Generates the storage path for a photo
func GeneratePhotoPath(benchID uint, photoID uint, size string) string {
	return fmt.Sprintf("benches/%d/photos/%d_%s.jpg", benchID, photoID, size)
}
