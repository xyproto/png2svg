package png2svg

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strconv"
	"strings"

	"github.com/xyproto/onthefly"
)

const VersionString = "1.1.0"

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
// Returns an empty string if no fill color is found
func colorFromLine(line string) string {
	if !strings.Contains(line, " fill=\"") {
		return ""
	}
	fields := strings.Fields(line)
	for _, field := range fields {
		if strings.HasPrefix(field, "fill=") {
			elems := strings.Split(field, "\"")
			// if len(elems) < 3 { panic("invalid SVG: " + line) }
			return elems[1]
		}
	}
	// Only used during debugging
	// panic("No fill color on this line: " + line)
	return ""
}

type Rect struct {
	x int
	y int
	w int
	h int
}

func (r *Rect) String() string {
	return fmt.Sprintf("<rect x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" />", r.x, r.y, r.w, r.h)
}

// adjacent returns true if two rectangles are adjacent to each other
func (r1 *Rect) adjacent(r2 *Rect) bool {
	if (r1.x + r1.w) == r2.x { // r1 is directly left of r2, if y or (y+h) is within range
		if ((r1.y > r2.y) && (r1.y <= (r2.y + r2.h))) || (((r1.y + r1.h) > r2.y) && ((r1.y + r1.h) <= (r2.y + r2.h))) {
			return true
		}
	}
	if (r1.y + r1.h) == r2.y { // r1 is directly above r2, if x or (x+w) is within range
		if ((r1.x > r2.x) && (r1.x <= (r2.x + r2.w))) || (((r1.x + r1.w) > r2.x) && ((r1.x + r1.w) <= (r2.x + r2.w))) {
			return true
		}
	}
	if (r2.x + r2.w) == r1.x { // r1 is directly right of r2, if y or (y+h) is within range
		if ((r1.y > r2.y) && (r1.y <= (r2.y + r2.h))) || (((r1.y + r1.h) > r2.y) && ((r1.y + r1.h) <= (r2.y + r2.h))) {
			return true
		}
	}
	if (r2.y + r2.h) == r1.y { // r1 is directly below r2, if x is within range
		if ((r1.x > r2.x) && (r1.x <= (r2.x + r2.w))) || (((r1.x + r1.w) > r2.x) && ((r1.x + r1.w) <= (r2.x + r2.w))) {
			return true
		}
	}
	return false
}

// allAdjacent returns all adjacent rectangles in a slice of rectangles
func (r *Rect) allAdjacent(rects []*Rect) []*Rect {
	var adj []*Rect
	adj = append(adj, r)
	for _, e := range rects {
		if e == r {
			continue
		}
		if r.adjacent(e) {
			adj = append(adj, e)
		}
	}
	if len(adj) < 2 {
		return []*Rect{}
	}
	return adj
}

// Check if the given map contains a given element,
// either in the keys or in one of the value slices.
func inValues(a map[*Rect][]*Rect, e *Rect) bool {
	for key, values := range a {
		if key == e {
			return true
		}
		for _, r := range values {
			if e == r {
				return true
			}
		}
	}
	return false
}

var COUNTER = 0

func rects2polygon(adjacentRects []*Rect) string {
	//fmt.Println("TO COMBINE:", adjacentRects)
	pc := NewPointCollection() // len(adjacentRects)*4
	var x, y float64
	for _, r := range adjacentRects {
		x = float64(r.x)
		y = float64(r.y)
		if !pc.HasXY(x, y) {
			pc.PushXY(x, y)
		}
		x = float64(r.x + r.w)
		y = float64(r.y)
		if !pc.HasXY(x, y) {
			pc.PushXY(x, y)
		}
		x = float64(r.x)
		y = float64(r.y + r.h)
		if !pc.HasXY(x, y) {
			pc.PushXY(x, y)
		}
		x = float64(r.x + r.w)
		y = float64(r.y + r.h)
		if !pc.HasXY(x, y) {
			pc.PushXY(x, y)
		}
	}

	fmt.Println(pc)
	fmt.Println("OK 1")

	// Now sort the points in clockwise order
	err := pc.GrahamScan()
	if err != nil {
		fmt.Println("OH NOES", err)
		panic(err)
	}

	fmt.Println("OK 2")

	pal := []string{"#571845", "#900c3e", "#c70039", "#ff5733", "#ffc300"}

	fmt.Println("POINTS", pc.String())
	fmt.Println("POLYGONSTRING", pc.PolygonString())
	COUNTER++
	if COUNTER%2 == 0 {
		return fmt.Sprintf("<polygon points=\"%s\" fill=\""+pal[0]+"\" />", pc.PolygonString())
	}
	if COUNTER%3 == 0 {
		return fmt.Sprintf("<polygon points=\"%s\" fill=\""+pal[1]+"\" />", pc.PolygonString())
	}
	if COUNTER%4 == 0 {
		return fmt.Sprintf("<polygon points=\"%s\" fill=\""+pal[2]+"\" />", pc.PolygonString())
	}
	if COUNTER%5 == 0 {
		return fmt.Sprintf("<polygon points=\"%s\" fill=\""+pal[3]+"\" />", pc.PolygonString())
	}
	return fmt.Sprintf("<polygon points=\"%s\" fill=\""+pal[4]+"\" />", pc.PolygonString())
	//return fmt.Sprintf("<polygon points=\"%s\" />", pc.PolygonString())
}

// This is an inefficient prototype of a way to combine rectangles to polygons
func combineRectangles(lines []string) []string {
	var rects []*Rect

	fmt.Println("GROUP OF RECTANGLES")

	for _, line := range lines {
		var r Rect
		//fmt.Println(line)
		fields := strings.Fields(line)
		for _, field := range fields {
			keyValue := strings.Split(field, "=")
			if len(keyValue) != 2 {
				continue
			}
			key, value := keyValue[0], keyValue[1]
			if value[0] != '"' {
				continue
			}
			switch key {
			case "x":
				i, err := strconv.Atoi(value[1 : len(value)-1])
				if err != nil {
					fmt.Println(err)
					continue
				}
				r.x = i
			case "y":
				i, err := strconv.Atoi(value[1 : len(value)-1])
				if err != nil {
					fmt.Println(err)
					continue
				}
				r.y = i
			case "width":
				i, err := strconv.Atoi(value[1 : len(value)-1])
				if err != nil {
					fmt.Println(err)
					continue
				}
				r.w = i
			case "height":
				i, err := strconv.Atoi(value[1 : len(value)-1])
				if err != nil {
					fmt.Println(err)
					continue
				}
				r.h = i
			}
		}
		//fmt.Println("RECT", r)
		rects = append(rects, &r)
	}

	// map of adjacent rectangles, with the first found adjacent rectangle as the key
	a := make(map[*Rect][]*Rect)

	for _, r1 := range rects {
		for _, r2 := range rects {
			if r1 == r2 {
				continue
			}
			if inValues(a, r1) {
				continue
			}
			if inValues(a, r2) {
				continue
			}
			if r1.adjacent(r2) {
				if a[r1] == nil {
					a[r1] = make([]*Rect, 0)
					a[r1] = append(a[r1], r1)
					a[r1] = append(a[r1], r2)
				} else {
					a[r1] = append(a[r1], r2)
				}
				//fmt.Println("ADJACENT GROUP NAME", r1, "CONTENTS", a[r1])
			}
		}
	}

	var rendered []string
	for _, r := range rects {
		if adjacentRects, ok := a[r]; ok {
			rendered = append(rendered, rects2polygon(adjacentRects))
			continue
		}
		if inValues(a, r) {
			// Already covered
			continue
		}
		rendered = append(rendered, r.String())
	}

	//fmt.Println(strings.Join(rendered, "\n\t"))

	return rendered
}

// group lines that has a fill color by color, and organize under <g> tags
func groupLinesByFillColor(lines []string) []string {
	// Group lines by fill color
	groupedLines := make(map[string][]string)
	var fillColor string
	for i, line := range lines {
		fillColor = colorFromLine(line)
		if fillColor == "" {
			continue
		}
		// Erase the line. The grouped lines will be inserted at the first empty line.
		lines[i] = ""
		if groupedLines[fillColor] == nil {
			groupedLines[fillColor] = make([]string, 0)
		}
		groupedLines[fillColor] = append(groupedLines[fillColor], line)
	}

	// Build a string of all lines with fillcolor, grouped by fillcolor, inside <g> tags
	var sb strings.Builder
	for fillColor, lines := range groupedLines {
		if len(lines) > 1 {
			// Combine rectangles into polygons, for the same fill color.
			// Also removes the fill color. The fillColor replace below can be skipped for this case.
			lines = combineRectangles(lines)
		}
		sb.WriteString("<g fill=\"" + fillColor + "\">")
		for _, line := range lines {
			if strings.Contains(line, "<rect") {
				rectString := strings.Replace(line, " fill=\""+fillColor+"\"", "", 1)
				sb.WriteString(rectString)
			} else {
				sb.WriteString(line)
			}
		}
		sb.WriteString("</g>")
	}
	contents := sb.String()

	// Insert the contents in the slice of lines
	inserted := false
	for i, line := range lines {
		if !inserted && line == "" {
			lines[i] = contents
			inserted = true
		}
	}

	// Return lines, some of them empty. One of them is a really long line with the above contents.
	return lines
}

// String returns the rendered SVG document as a string
func (pi *PixelImage) String() string {
	if pi.verbose {
		fmt.Print("Rendering SVG...")
	}

	// Render the SVG document
	svgDocument := pi.page.String()

	if pi.verbose {
		fmt.Println("ok")
	}

	if pi.verbose {
		fmt.Print("Grouping elements by color...")
	}

	// Group lines by fill color, insert <g> tags
	lines := groupLinesByFillColor(strings.Split(svgDocument, "\n"))

	// Use the new line contents as the new svgDocument
	svgDocument = strings.Join(lines, "\n")

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
	svgDocument = strings.Replace(svgDocument, "\n", "", -1)
	// Remove all spaces before closing tags
	svgDocument = strings.Replace(svgDocument, " />", "/>", -1)
	// Remove double spaces
	svgDocument = strings.Replace(svgDocument, "  ", " ", -1)
	// Remove empty x attributes
	svgDocument = strings.Replace(svgDocument, " x=\"0\"", "", -1)
	// Remove empty y attributes
	svgDocument = strings.Replace(svgDocument, " y=\"0\"", "", -1)
	// Remove empty width attributes
	svgDocument = strings.Replace(svgDocument, " width=\"0\"", "", -1)
	// Remove empty height attributes
	svgDocument = strings.Replace(svgDocument, " height=\"0\"", "", -1)
	// Remove single spaces between tags
	svgDocument = strings.Replace(svgDocument, "> <", "><", -1)
	// "red" is shorter than #f00 or #ff0000
	svgDocument = strings.Replace(svgDocument, "#f00", "red", -1)
	svgDocument = strings.Replace(svgDocument, "#ff0000", "red", -1)
	// "white" is shorter than #ffffff
	svgDocument = strings.Replace(svgDocument, "#ffffff", "white", -1)
	// "black" is shorter than #000000
	svgDocument = strings.Replace(svgDocument, "#000000", "black", -1)

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
	if _, err = f.WriteString(pi.String()); err != nil {
		return err
	}
	return nil
}
