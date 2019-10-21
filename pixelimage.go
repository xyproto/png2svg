package png2svg

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"

	"github.com/xyproto/tinysvg"
)

const VersionString = "1.2.0"

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
	pixels   Pixels
	document *tinysvg.Document
	svgTag   *tinysvg.Tag
	verbose  bool
	w        int
	h        int
}

func ReadPNG(filename string, verbose bool) (image.Image, error) {
	if verbose {
		fmt.Printf("Reading %s\n", filename)
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
	return img, nil
}

// Erase characters on the terminal
func Erase(n int) {
	fmt.Print(strings.Repeat("\b", n))
}

func NewPixelImage(img image.Image, verbose bool) *PixelImage {
	width := img.Bounds().Max.X - img.Bounds().Min.X
	height := img.Bounds().Max.Y - img.Bounds().Min.Y

	pixels := make(Pixels, width*height, width*height)

	var c color.NRGBA
	if verbose {
		fmt.Print("Interpreting image... 0%")
	}

	percentage := 0
	lastPercentage := 0
	i := 0
	lastLine := img.Bounds().Max.Y

	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {

		if verbose && y != lastLine {
			lastPercentage = percentage
			percentage = int((float64(y) / float64(height)) * 100.0)
			Erase(len(fmt.Sprintf("%d%%", lastPercentage)))
			fmt.Printf("%d%%", percentage)
			lastLine = y
		}

		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			c = color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			alpha := int(c.A)
			// Mark transparent pixels as already being "covered"
			covered := alpha == 0
			pixels[i] = &Pixel{x, y, int(c.R), int(c.G), int(c.B), alpha, covered}
			i++
		}
	}

	// Create a new XML document with a new SVG tag
	document, svgTag := tinysvg.NewTinySVG(width, height)

	if verbose {
		Erase(len(fmt.Sprintf("%d%%", lastPercentage)))
		fmt.Println("100%")
	}

	return &PixelImage{pixels, document, svgTag, verbose, width, height}
}

// Done checks if all pixels are covered, in terms of being represented by an SVG element
// searches from the given x and y coordinate
func (pi *PixelImage) Done(startx, starty int) bool {
	for y := starty; y < pi.h; y++ {
		for x := startx; x < pi.w; x++ {
			i := y*pi.w + x
			if !pi.pixels[i].covered {
				return false
			}
		}
		// Start at the beginning of the line when searching the rest of the lines
		startx = 0
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

func (pi *PixelImage) FirstUncovered(startx, starty int) (int, int) {
	for y := starty; y < pi.h; y++ {
		for x := startx; x < pi.w; x++ {
			i := y*pi.w + x
			if !pi.pixels[i].covered {
				return x, y
			}
		}
		// Start at the beginning of the line when searching the rest of the lines
		startx = 0
	}
	// This should never happen, except when debugging
	panic("All pixels are covered")
}

// Extract the fill color from a svg rect line (<rect ... fill="#ff0000" ...) would return #ff0000
// Returns false if no fill color is found
// Returns an empty string if no fill color is found
func colorFromLine(line []byte) ([]byte, bool) {
	if !bytes.Contains(line, []byte(" fill=\"")) {
		return nil, false
	}
	fields := bytes.Fields(line)
	for _, field := range fields {
		if bytes.HasPrefix(field, []byte("fill=")) {
			elems := bytes.Split(field, []byte("\""))
			// if len(elems) < 3 { panic("invalid SVG: " + line) }
			// assumption: if fill= exists, elems[1] exists
			return elems[1], true
		}
	}
	// This should never happen
	return nil, false
}

// groupLinesByFillColor will group lines that has a fill color by color, organized under <g> tags
// This is not the prettiest function, but it works.
// TODO: Rewrite, to make it prettier
// TODO: Benchmark
func groupLinesByFillColor(lines [][]byte) [][]byte {
	// Group lines by fill color
	var (
		groupedLines = make(map[string][][]byte)
		fillColor    []byte
		found        bool
	)
	for i, line := range lines {
		fillColor, found = colorFromLine(line)
		if !found {
			// skip
			continue
		}
		// Erase this line. The grouped lines will be inserted at the first empty line.
		lines[i] = make([]byte, 0)
		cs := string(fillColor)
		if _, ok := groupedLines[cs]; !ok {
			// Start an empty line
			groupedLines[cs] = make([][]byte, 0)
		}
		line = append(line, '>')
		//fmt.Println("ADDING", string(line), "TO LINE AT KEY", cs)
		groupedLines[cs] = append(groupedLines[cs], line)
	}

	// Build a string of all lines with fillcolor, grouped by fillcolor, inside <g> tags
	var (
		buf  bytes.Buffer
		from []byte
	)
	for key, lines := range groupedLines {
		if len(lines) > 1 {
			buf.Write([]byte("<g fill=\""))
			buf.WriteString(key)
			buf.Write([]byte("\">"))
			for _, line := range lines {
				from = append([]byte(" fill=\""), key...)
				buf.Write(bytes.Replace(line, append(from, '"'), []byte{}, 1))
			}
			buf.Write([]byte("</g>"))
		} else {
			buf.Write(lines[0])
		}
	}
	// Insert the contents in the first non-empty slice of lines
	for i, line := range lines {
		if len(line) == 0 {
			lines[i] = buf.Bytes()
			break
		}
	}
	// Return lines, some of them empty. One of them is a really long line with the above contents.
	return lines
}

// Bytes returns the rendered SVG document as bytes
func (pi *PixelImage) Bytes() []byte {
	if pi.verbose {
		fmt.Print("Rendering SVG...")
	}

	// Render the SVG document
	svgDocument := pi.document.Bytes()

	if pi.verbose {
		fmt.Println("ok")
	}

	if pi.verbose {
		fmt.Print("Grouping elements by color...")
	}

	// TODO: Make the code related to grouping both faster and more readable

	// Group lines by fill color, insert <g> tags
	lines := bytes.Split(svgDocument, []byte(">"))
	lines = groupLinesByFillColor(lines)

	// Use the new line contents as the new svgDocument
	for i, line := range lines {
		if len(line) > 0 && !bytes.HasSuffix(line, []byte(">")) {
			lines[i] = append(line, '>')
		}
	}
	svgDocument = bytes.Join(lines, []byte{})

	//fmt.Println("SVG DOCUMENT AFTER GROUPING")
	//fmt.Println(string(svgDocument))

	if pi.verbose {
		fmt.Println("ok")
	}

	// Only non-destructive and spec-conforming optimizations goes here

	// NOTE: Removing width and height for "1" gave incorrect results in GIMP.
	// NOTE: Gimp complains about the width and height not being set, but it is set.

	if pi.verbose {
		fmt.Print("Additional optimizations...")
	}

	// Remove all newlines
	// Remove all spaces before closing tags
	// Remove double spaces
	// Remove empty x attributes
	// Remove empty y attributes
	// Remove empty width attributes
	// Remove empty height attributes
	// Remove single spaces between tags
	// "red" is shorter than #f00 or #ff0000
	// "white" is shorter than #ffffff
	// "black" is shorter than #000000
	svgDocument = bytes.Replace(svgDocument, []byte("\n"), []byte(""), -1)
	svgDocument = bytes.Replace(svgDocument, []byte(" />"), []byte("/>"), -1)
	svgDocument = bytes.Replace(svgDocument, []byte("  "), []byte(" "), -1)
	svgDocument = bytes.Replace(svgDocument, []byte(" x=\"0\""), []byte(""), -1)
	svgDocument = bytes.Replace(svgDocument, []byte(" y=\"0\""), []byte(""), -1)
	svgDocument = bytes.Replace(svgDocument, []byte(" width=\"0\""), []byte(""), -1)
	svgDocument = bytes.Replace(svgDocument, []byte(" height=\"0\""), []byte(""), -1)
	svgDocument = bytes.Replace(svgDocument, []byte("> <"), []byte("><"), -1)
	svgDocument = bytes.Replace(svgDocument, []byte("#f00"), []byte("red"), -1)
	svgDocument = bytes.Replace(svgDocument, []byte("#ff0000"), []byte("red"), -1)
	svgDocument = bytes.Replace(svgDocument, []byte("#ffffff"), []byte("#fff"), -1)
	svgDocument = bytes.Replace(svgDocument, []byte("#000000"), []byte("#000"), -1)

	if pi.verbose {
		fmt.Println("ok")
	}

	return svgDocument
}

func (pi *PixelImage) WriteSVG(filename string) error {
	var (
		err error
		f   *os.File
	)

	if !pi.Done(0, 0) {
		return errors.New("the SVG representation does not cover all pixels")
	}
	if filename == "-" {
		f = os.Stdout
		// Turn off verbose messages, so that they don't end up in the SVG output
		pi.verbose = false
	} else {
		f, err = os.Create(filename)
		if err != nil {
			return err
		}
		defer f.Close()
	}

	// Write the generated SVG image to file or to stdout
	if _, err = f.Write(pi.Bytes()); err != nil {
		return err
	}
	return nil
}
