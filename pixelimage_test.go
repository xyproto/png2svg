package png2svg

import (
	"testing"
)

func TestCorrectPixelDataAt10_1(t *testing.T) {
	const filename = "img/rainforest.png"
	const targetX, targetY = 10, 1
	const expectedRed, expectedGreen, expectedBlue = 0, 0, 0 // Assuming the pixel should be black

	// Read the image
	img, err := ReadPNG(filename, false)
	if err != nil {
		t.Fatalf("Failed to read PNG file: %v", err)
	}

	// Initialize PixelImage
	pixelImage := NewPixelImage(img, false)

	// Get the pixel at (10, 1)
	pixel := pixelImage.pixels[targetY*pixelImage.w+targetX]

	// Check if the color of the pixel matches the expected values
	if pixel.r != expectedRed || pixel.g != expectedGreen || pixel.b != expectedBlue {
		t.Errorf("Pixel at (%d,%d) has incorrect color: got (R: %d, G: %d, B: %d), want (R: %d, G: %d, B: %d)",
			targetX, targetY, pixel.r, pixel.g, pixel.b, expectedRed, expectedGreen, expectedBlue)
	}
}
