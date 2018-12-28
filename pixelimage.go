package png2svg

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"

	"github.com/xyproto/onthefly"
)

type Pixel struct {
	x       int
	y       int
	r       int
	g       int
	b       int
	a       int
	covered bool
}

type Pixels []*Pixel

type PixelImage struct {
	pixels  Pixels
	page    *onthefly.Page
	svgTag  *onthefly.Tag
	verbose bool
	w       int
	h       int
}

func ReadPNG(filename string, verbose bool) (image.Image, error) {
	if verbose {
		fmt.Printf("Reading %s...", filename)
	}
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	if verbose {
		fmt.Println("ok")
	}
	return img, nil
}

func NewPixelImage(img image.Image, verbose bool) *PixelImage {
	width := img.Bounds().Max.X - img.Bounds().Min.X
	height := img.Bounds().Max.Y - img.Bounds().Min.Y

	pixels := make(Pixels, width*height, width*height)

	var c color.NRGBA
	if verbose {
		fmt.Print("Converting image.Image to Pixels")
	}
	i := 0
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		if verbose {
			fmt.Print(".")
		}
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			c = color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			alpha := int(c.A)
			// Mark transparent pixels as already being "covered"
			covered := alpha == 0
			//if covered {
			//	fmt.Println("ALPHA AT", x, y)
			//}
			pixels[i] = &Pixel{x, y, int(c.R), int(c.G), int(c.B), alpha, covered}
			i++
		}
	}

	// Create a new XML page with a new SVG tag
	page, svgTag := onthefly.NewTinySVGPixels(width, height)

	if verbose {
		fmt.Println("ok")
	}

	return &PixelImage{pixels, page, svgTag, verbose, width, height}
}

// Done checks if all pixels are covered, in terms of being represented by an SVG element
func (pi *PixelImage) Done() bool {
	for _, p := range pi.pixels {
		if !(*p).covered {
			return false
		}
	}
	return true
}

// At returns the RGB color at the given coordinate
func (pi *PixelImage) At(x, y int) (r, g, b int) {
	i := y*pi.w + x
	//if i >= len(pi.pixels) {
	//	panic("At out of bounds, too large coordinate")
	//}
	p := *pi.pixels[i]
	return p.r, p.g, p.b
}

// Covered returns true if the pixel at the given coordinate is already covered by SVG elements
func (pi *PixelImage) Covered(x, y int) bool {
	i := y*pi.w + x
	p := *pi.pixels[i]
	return p.covered
}

// Cover all pixels that are not yet covered, by creating an svg rectangle per pixel
func (pi *PixelImage) CoverAllPixels() {
	coverCount := 0
	for _, p := range pi.pixels {
		if !(*p).covered {
			pi.svgTag.Pixel((*p).x, (*p).y, (*p).r, (*p).g, (*p).b)
			(*p).covered = true
			coverCount++
		}
	}
	if pi.verbose {
		fmt.Printf("Covered %d pixels with 1x1 rectangles.\n", coverCount)
	}
}

func (pi *PixelImage) FirstUncovered() (int, int) {
	for y := 0; y < pi.h; y++ {
		for x := 0; x < pi.w; x++ {
			i := y*pi.w + x
			if !pi.pixels[i].covered {
				return x, y
			}
		}
	}
	panic("All pixels are covered")
}

func (pi *PixelImage) WriteSVG(filename string, optimize bool) error {
	if !pi.Done() {
		return errors.New("the SVG representation does not cover all pixels")
	}
	if pi.verbose {
		fmt.Printf("Writing %s...", filename)
	}
	var (
		err error
		f   *os.File
	)
	if filename == "-" {
		f = os.Stdout
	} else {
		f, err = os.Create(filename)
		if err != nil {
			return err
		}
		defer f.Close()
	}
	svgDocument := pi.page.String()

	if optimize {
		// Remove all newlines
		svgDocument = strings.Replace(svgDocument, "\n", "", -1)
		// Remove all spaces before closing tags
		svgDocument = strings.Replace(svgDocument, " />", "/>", -1)
		// NOTE: Removing width and height for "1" gave incorrect results in GIMP.
		// TODO: Remove quotes around rectangle x/y/width/height?
	}

	if _, err = f.WriteString(svgDocument); err != nil {
		return err
	}
	if pi.verbose {
		fmt.Println("ok")
	}
	return nil
}
