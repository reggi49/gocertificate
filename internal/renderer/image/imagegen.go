package imagegen

import (
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
)

// GenerateTestPNGWithAlpha creates a sample PNG with transparent background and colored shape.
func GenerateTestPNGWithAlpha(filePath string, width, height int, col color.RGBA) error {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with transparent
	draw.Draw(img, img.Bounds(), &image.Uniform{color.Transparent}, image.Point{}, draw.Src)

	// Draw an inner colored rectangle with alpha
	for y := height / 4; y < (3*height)/4; y++ {
		for x := width / 4; x < (3*width)/4; x++ {
			img.Set(x, y, col)
		}
	}

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}

// GenerateTestJPEG creates a solid background JPEG image.
func GenerateTestJPEG(filePath string, width, height int, col color.RGBA) error {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{col}, image.Point{}, draw.Src)

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return jpeg.Encode(f, img, &jpeg.Options{Quality: 90})
}
