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

// Pixel represents a pixel at position (x,y)
// with color (r,g,b,a)
// and a bool for if this pixel has been covered by an SVG shape yet
type Pixel struct {
	x       int
	y       int
	r       int
	g       int
	b       int
	a       int
	covered bool
}

// Pixels is a slice of pointers to Pixel
type Pixels []*Pixel

// PixelImage contains the data needed to convert a PNG to an SVG:
// pixels (with an overview of which pixels are covered) and
// an SVG document, starting with the document and root tag +
// colorOptimize, for if only 4096 colors should be used
// (short hex color strings, like #fff).
type PixelImage struct {
	pixels        Pixels
	document      *tinysvg.Document
	svgTag        *tinysvg.Tag
	verbose       bool
	w             int
	h             int
	colorOptimize bool
}

// SetColorOptimize can be used to set the colorOptimize flag,
// for using only 4096 colors.
func (pi *PixelImage) SetColorOptimize(enabled bool) {
	pi.colorOptimize = enabled
}

// ReadPNG tries to read the given PNG image filename and returns and image.Image
// and an error. If verbose is true, some basic information is printed to stdout.
func ReadPNG(filename string, verbose bool) (image.Image, error) {
	if verbose {
		fmt.Printf("Reading %s", filename)
		defer fmt.Println()
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
		fmt.Printf(" (%dx%d)", img.Bounds().Max.X-img.Bounds().Min.X, img.Bounds().Max.Y-img.Bounds().Min.Y)
	}
	return img, nil
}

// Erase characters on the terminal
func Erase(n int) {
	fmt.Print(strings.Repeat("\b", n))
}

// NewPixelImage initializes a new PixelImage struct,
// given an image.Image.
func NewPixelImage(img image.Image, verbose bool) *PixelImage {
	width := img.Bounds().Max.X - img.Bounds().Min.X
	height := img.Bounds().Max.Y - img.Bounds().Min.Y

	pixels := make(Pixels, width*height)

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
			// Mark transparent pixels as already being "covered" (alpha == 0)
			pixels[i] = &Pixel{x, y, int(c.R), int(c.G), int(c.B), int(c.A), c.A == 0}
			i++
		}
	}

	// Create a new XML document with a new SVG tag
	document, svgTag := tinysvg.NewTinySVG(width, height)

	if verbose {
		Erase(len(fmt.Sprintf("%d%%", lastPercentage)))
		fmt.Println("100%")
	}

	return &PixelImage{pixels, document, svgTag, verbose, width, height, false}
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

// At2 returns the RGBA color at the given coordinate
func (pi *PixelImage) At2(x, y int) (r, g, b, a int) {
	i := y*pi.w + x
	//if i >= len(pi.pixels) {
	//	panic("At out of bounds, too large coordinate")
	//}
	p := *pi.pixels[i]
	return p.r, p.g, p.b, p.a
}

// Covered returns true if the pixel at the given coordinate is already covered by SVG elements
func (pi *PixelImage) Covered(x, y int) bool {
	i := y*pi.w + x
	p := *pi.pixels[i]
	return p.covered
}

// CoverAllPixels will cover all pixels that are not yet covered by an SVG element,
// by creating a rectangle per pixel.
func (pi *PixelImage) CoverAllPixels() {
	coverCount := 0
	for _, p := range pi.pixels {
		if !(*p).covered {
			pi.svgTag.Pixel((*p).x, (*p).y, int((*p).r), int((*p).g), int((*p).b))
			(*p).covered = true
			coverCount++
		}
	}
	if pi.verbose {
		fmt.Printf("Covered %d pixels with 1x1 rectangles.\n", coverCount)
	}
}

// CoverAllPixelsCallback will cover all pixels that are not yet covered by an SVG element,
// by creating a rectangle per pixel. Also takes a callback function that will be called
// with which pixel index the program is at and also the total pixels, for each Nth pixels (and at the start and end).
func (pi *PixelImage) CoverAllPixelsCallback(callbackFunc func(int, int), Nth int) {
	coverCount := 0
	l := len(pi.pixels)
	callbackFunc(0, l)
	for i, p := range pi.pixels {
		if !(*p).covered {
			pi.svgTag.Pixel((*p).x, (*p).y, int((*p).r), int((*p).g), int((*p).b))
			(*p).covered = true
			coverCount++
		}
		if i%Nth == 0 {
			callbackFunc(i, l)
		}
	}
	callbackFunc(l-1, l)
	if pi.verbose {
		fmt.Printf("Covered %d pixels with 1x1 rectangles.\n", coverCount)
	}
}

// FirstUncovered will find the first pixel that is not covered by an SVG element,
// starting from (startx,starty), searching row-wise, downwards.
func (pi *PixelImage) FirstUncovered(startx, starty int) (int, int) {
	for y := starty; y < pi.h; y++ {
		for x := startx; x < pi.w; x++ {
			if !pi.pixels[y*pi.w+x].covered {
				return x, y
			}
		}
		// Start at the beginning of the line when searching the rest of the lines
		startx = 0
	}
	// This should never happen, except when debugging
	panic("All pixels are covered")
}

// colorFromLine will extract the fill color from a svg rect line.
// "<rect ... fill="#ff0f00" ..." gives "#ff0f00".
// #ff0000 is shortened to  #f00.
// Returns false if no fill color is found.
// Returns an empty string if no fill color is found.
func colorFromLine(line []byte, lossyColorCompression bool) ([]byte, []byte, bool) {
	if !bytes.Contains(line, []byte(" fill=\"")) {
		return nil, nil, false
	}
	fields := bytes.Fields(line)
	var (
		fillEquals = []byte("fill=")
		bsq        = []byte("\"")
		elems      [][]byte
	)
	if lossyColorCompression {
		for _, field := range fields {
			if bytes.HasPrefix(field, fillEquals) {
				// assumption: there are always quotes, so that elems[1] exists
				elems = bytes.Split(field, bsq)
				return elems[1], shortenColorLossy(elems[1]), true
			}
		}
	} else {
		for _, field := range fields {
			if bytes.HasPrefix(field, fillEquals) {
				// assumption: there are always quotes, so that elems[1] exists
				elems = bytes.Split(field, bsq)
				return elems[1], shortenColorLossless(elems[1]), true
			}
		}
	}
	// This should never happen
	return nil, nil, false
}

// groupLinesByFillColor will group lines that has a fill color by color, organized under <g> tags
// This is not the prettiest function, but it works.
// TODO: Rewrite, to make it prettier
// TODO: Benchmark
func groupLinesByFillColor(lines [][]byte, colorOptimize bool) [][]byte {
	// Group lines by fill color
	var (
		groupedLines                  = make(map[string][][]byte)
		fillColor, shortenedFillColor []byte
		found                         bool
	)
	for i, line := range lines {
		fillColor, shortenedFillColor, found = colorFromLine(line, colorOptimize)
		if !found {
			// skip
			continue
		}
		// Erase this line. The grouped lines will be inserted at the first empty line.
		lines[i] = make([]byte, 0)
		// // Convert from []byte to string because map keys can't be []byte
		cs := string(shortenedFillColor)
		if _, ok := groupedLines[cs]; !ok {
			// Start an empty line
			groupedLines[cs] = make([][]byte, 0)
		}
		if string(fillColor) != cs {
			line = bytes.Replace(line, fillColor, shortenedFillColor, 1)
		}
		line = append(line, '>')
		groupedLines[cs] = append(groupedLines[cs], line)
	}

	//for k, _ := range groupedLines {
	//	fmt.Println("COLOR: ", string(k))
	//}

	// Build a string of all lines with fillcolor, grouped by fillcolor, inside <g> tags
	var (
		buf       bytes.Buffer
		from      []byte
		gfill     = []byte("<g fill=\"")
		closing   = []byte("\">")
		spacefill = []byte(" fill=\"")
	)
	for key, lines := range groupedLines {
		if len(lines) > 1 {
			buf.Write(gfill)
			//fmt.Printf("WRITING KEY %s\n", key)
			buf.WriteString(key)
			buf.Write(closing)
			for _, line := range lines {
				from = append(spacefill, key...)
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
	// TODO: pi.document.WriteTo also exists, and might be faster
	svgDocument := pi.document.Bytes()

	if pi.verbose {
		fmt.Println("ok")
		fmt.Print("Grouping elements by color...")
	}

	// TODO: Make the code related to grouping both faster and more readable

	// Group lines by fill color, insert <g> tags
	lines := bytes.Split(svgDocument, []byte(">"))
	lines = groupLinesByFillColor(lines, pi.colorOptimize)

	for i, line := range lines {
		if len(line) > 0 && !bytes.HasSuffix(line, []byte(">")) {
			lines[i] = append(line, '>')
		}
	}
	// Use the line contents as the new svgDocument
	svgDocument = bytes.Join(lines, []byte{})

	if pi.verbose {
		fmt.Println("ok")
		fmt.Print("Additional optimizations...")
	}

	// Only non-destructive and spec-conforming optimizations goes here

	// NOTE: Removing width and height for "1" gave incorrect results in GIMP.
	// NOTE: GIMP complains about the width and height not being set, but it is set.

	// Remove all newlines
	// Remove all spaces before closing tags
	// Remove double spaces
	// Remove empty x attributes
	// Remove empty y attributes
	// Remove empty width attributes
	// Remove empty height attributes
	// Remove single spaces between tags

	svgDocument = bytes.Replace(svgDocument, []byte("\n"), []byte{}, -1)
	svgDocument = bytes.Replace(svgDocument, []byte(" />"), []byte("/>"), -1)
	svgDocument = bytes.Replace(svgDocument, []byte("  "), []byte(" "), -1)
	svgDocument = bytes.Replace(svgDocument, []byte(" x=\"0\""), []byte{}, -1)
	svgDocument = bytes.Replace(svgDocument, []byte(" y=\"0\""), []byte{}, -1)
	svgDocument = bytes.Replace(svgDocument, []byte(" width=\"0\""), []byte{}, -1)
	svgDocument = bytes.Replace(svgDocument, []byte(" height=\"0\""), []byte{}, -1)
	svgDocument = bytes.Replace(svgDocument, []byte("> <"), []byte("><"), -1)

	// Replacement of colors that are not shortened, colors that has been shortened
	// and color names to even shorter strings.
	colorReplacements := map[string][]byte{
		"#f0ffff": []byte("azure"),
		"#f5f5dc": []byte("beige"),
		"#ffe4c4": []byte("bisque"),
		"#a52a2a": []byte("brown"),
		"#ff7f50": []byte("coral"),
		"#ffd700": []byte("gold"),
		"#808080": []byte("gray"), // "grey" is also possible
		"#008000": []byte("green"),
		"#4b0082": []byte("indigo"),
		"#fffff0": []byte("ivory"),
		"#f0e68c": []byte("khaki"),
		"#faf0e6": []byte("linen"),
		"#800000": []byte("maroon"),
		"#000080": []byte("navy"),
		"#808000": []byte("olive"),
		"#ffa500": []byte("orange"),
		"#da70d6": []byte("orchid"),
		"#cd853f": []byte("peru"),
		"#ffc0cb": []byte("pink"),
		"#dda0dd": []byte("plum"),
		"#800080": []byte("purple"),
		"#f00":    []byte("red"),
		"#fa8072": []byte("salmon"),
		"#a0522d": []byte("sienna"),
		"#c0c0c0": []byte("silver"),
		"#fffafa": []byte("snow"),
		"#d2b48c": []byte("tan"),
		"#008080": []byte("teal"),
		"#ff6347": []byte("tomato"),
		"#ee82ee": []byte("violet"),
		"#f5deb3": []byte("wheat"),
	}

	// Replace colors with the shorter version
	for k, v := range colorReplacements {
		svgDocument = bytes.Replace(svgDocument, []byte(k), v, -1)
	}

	if pi.verbose {
		fmt.Println("ok")
	}

	return svgDocument
}

// WriteSVG will save the current SVG document to a file
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
