package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"os"
	"time"

	"github.com/xyproto/onthefly"
)

type Pixel struct {
	x       int
	y       int
	r       int
	g       int
	b       int
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
			pixels[i] = &Pixel{x, y, int(c.R), int(c.G), int(c.B), false}
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

// Check if all pixels are covered, in terms of being represented by an SVG element
func (pi *PixelImage) Done() bool {
	for _, p := range pi.pixels {
		if !(*p).covered {
			return false
		}
	}
	return true
}

type Box struct {
	x int
	y int
	w int
	h int
	r int
	g int
	b int
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func (pi *PixelImage) At(x, y int) (r, g, b int) {
	i := y*pi.w + x
	if i >= len(pi.pixels) {
		panic("At out of bounds, too large coordinate")
	}
	p := *pi.pixels[i]
	return p.r, p.g, p.b
}

func (pi *PixelImage) Covered(x, y int) bool {
	i := y*pi.w + x
	p := *pi.pixels[i]
	return p.covered
}

func (pi *PixelImage) CreateRandomBox(checkIfPossible bool) *Box {
	w := 1
	h := 1
	var x, y, r, g, b int
	for !checkIfPossible || !pi.Done() {
		// Find a random placement for (x,y), for a box of size (1,1)
		x = rand.Intn(pi.w)
		y = rand.Intn(pi.h)
		fmt.Printf("Random box a (%d, %d)\n", x, y)
		if pi.Covered(x, y) {
			continue
		}
		r, g, b = pi.At(x, y)
		break
	}
	// Create a box at that placement, with width 1 and height 1
	// Return the box
	return &Box{x, y, w, h, r, g, b}
}

// Expand a box to the left, if all new pixels have the same color
func (pi *PixelImage) ExpandLeft(bo *Box) bool {
	ok := true
	// Loop from box top left (-1,0) to box bot left (-1,0)
	x := bo.x - 1
	if x <= 0 {
		return false
	}
	for y := bo.y; y < (bo.y + bo.h); y++ {
		fmt.Printf("Expand left at (%d, %d)\n", x, y)
		r, g, b := pi.At(x, y)
		if !(r == bo.r && g == bo.g && b == bo.b) {
			ok = false
			break
		}
	}
	// Expand the box 1 pixel to the left
	bo.w++
	bo.x--

	fmt.Println("EXPANDED AT ", bo.x, bo.y)
	return ok
}

// Expand a box to the right, if all new pixels have the same color
func (pi *PixelImage) ExpandRight(bo *Box) bool {
	ok := true
	// Loop from box top right (+1,0) to box bot right (+1,0)
	x := bo.x + bo.w + 1
	if x >= pi.w {
		return false
	}
	for y := bo.y; y < (bo.y + bo.h); y++ {
		fmt.Printf("Expand right at (%d, %d)\n", x, y)
		r, g, b := pi.At(x, y)
		if !(r == bo.r && g == bo.g && b == bo.b) {
			ok = false
			break
		}
	}
	// Expand the box 1 pixel to the right
	bo.w++

	fmt.Println("EXPANDED AT ", bo.x, bo.y)
	return ok
}

func (pi *PixelImage) CoverBox(bo *Box, pink bool) {
	coverCount := 0
	// Draw the rectangle
	rect := pi.svgTag.AddRect(bo.x, bo.y, bo.w, bo.h)
	if pink {
		rect.Fill(onthefly.ColorString(0xbb, 0x33, 0x85))
	} else {
		rect.Fill(onthefly.ColorString(bo.r, bo.g, bo.b))
	}
	// Mark all covered pixels in the PixelImage
	for y := bo.y; y < (bo.y + bo.h); y++ {
		for x := bo.x; x < (bo.x + bo.w); x++ {
			i := y*pi.w + x
			pi.pixels[i].covered = true
			coverCount++
		}
	}
	if pi.verbose {
		fmt.Printf("Covered %d pixels with a custom rectangle.\n", coverCount)
	}
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

func (pi *PixelImage) WriteSVG(filename string) error {
	if !pi.Done() {
		return errors.New("the SVG representation does not cover all pixels")
	}
	if pi.verbose {
		fmt.Printf("Writing %s...", filename)
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(pi.page.String())
	if err != nil {
		return err
	}
	if pi.verbose {
		fmt.Println("ok")
	}
	return nil

}

func main() {
	var outputFilename string
	var colorPink bool
	flag.StringVar(&outputFilename, "o", "output.svg", "output SVG filename")
	flag.BoolVar(&colorPink, "p", false, "color expanded rectangles pink")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "%s\n%s\n",
			"An input PNG filename is required.",
			"Try: ./png2svg -o output.svg input.png")
		os.Exit(1)
	}

	inputFilename := args[0]
	verbose := true

	img, err := ReadPNG(inputFilename, verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	pi := NewPixelImage(img, verbose)

	randomBoxesToCreate := 2
	for !pi.Done() {
		bo := pi.CreateRandomBox(false)
		for pi.ExpandLeft(bo) {
		}
		for pi.ExpandRight(bo) {
		}
		pi.CoverBox(bo, colorPink)
		randomBoxesToCreate--
		if randomBoxesToCreate == 0 {
			break
		}
	}

	pi.CoverAllPixels()

	err = pi.WriteSVG(outputFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
