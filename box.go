package png2svg

import (
	"fmt"
	"math/rand"
	"strconv"
)

// Box represents a box with the following properties:
// * position (x, y)
// * size (w, h)
// * color (r, g, b, a)
type Box struct {
	x, y       int
	w, h       int
	r, g, b, a int
}

// CreateRandomBox randomly searches for a place for a 1x1 size box.
// Note: If checkIfPossible is true, the function continue running until
// it either finds a free spot or no spots are available.
func (pi *PixelImage) CreateRandomBox(checkIfPossible bool) *Box {
	w := 1
	h := 1
	var x, y, r, g, b, a int
	for !checkIfPossible || !pi.Done(0, 0) {
		// Find a random placement for (x,y), for a box of size (1,1)
		x = rand.Intn(pi.w)
		y = rand.Intn(pi.h)
		if pi.verbose {
			fmt.Printf("Random box at (%d, %d)\n", x, y)
		}
		if pi.Covered(x, y) {
			continue
		}
		r, g, b, a = pi.At2(x, y)
		break
	}
	// Create a box at that placement, with width 1 and height 1
	// Return the box
	return &Box{x, y, w, h, r, g, b, a}
}

// CreateBox creates a 1x1 box at the given location, if it's not already covered
func (pi *PixelImage) CreateBox(x, y int) *Box {
	if pi.Covered(x, y) {
		panic("CreateBox at location that was already covered")
	}
	w, h := 1, 1
	r, g, b, a := pi.At2(x, y)
	// Create a box at that placement, with width 1 and height 1
	// Return the box
	return &Box{x, y, w, h, r, g, b, a}
}

// ExpandLeft will expand a box 1 pixel to the left,
// if all new pixels have the same color
func (pi *PixelImage) ExpandLeft(bo *Box) bool {
	// Loop from box top left (-1,0) to box bot left (-1,0)
	x := bo.x - 1
	if x <= 0 {
		return false
	}
	for y := bo.y; y < (bo.y + bo.h); y++ {
		r, g, b, a := pi.At2(x, y)
		if (r != bo.r) || (g != bo.g) || (b != bo.b) || (a != bo.a) {
			return false
		}
	}
	// Expand the box 1 pixel to the left
	bo.w++
	bo.x--
	return true
}

// ExpandUp will expand a box 1 pixel upwards,
// if all new pixels have the same color
func (pi *PixelImage) ExpandUp(bo *Box) bool {
	// Loop from box top left to box top right
	y := bo.y - 1
	if y <= 0 {
		return false
	}
	for x := bo.x; x < (bo.x + bo.w); x++ {
		r, g, b, a := pi.At2(x, y)
		if (r != bo.r) || (g != bo.g) || (b != bo.b) || (a != bo.a) {
			return false
		}
	}
	// Expand the box 1 pixel up
	bo.h++
	bo.y--
	return true
}

// ExpandRight will expand a box 1 pixel to the right,
// if all new pixels have the same color
func (pi *PixelImage) ExpandRight(bo *Box) bool {
	// Loop from box top right (+1,0) to box bot right (+1,0)
	x := bo.x + bo.w //+ 1
	if x >= pi.w {
		return false
	}
	for y := bo.y; y < (bo.y + bo.h); y++ {
		r, g, b, a := pi.At2(x, y)
		if (r != bo.r) || (g != bo.g) || (b != bo.b) || (a != bo.a) {
			return false
		}
	}
	// Expand the box 1 pixel to the right
	bo.w++
	return true
}

// ExpandDown will expand a box 1 pixel downwards,
// if all new pixels have the same color
func (pi *PixelImage) ExpandDown(bo *Box) bool {
	// Loop from box bot left to box bot right
	y := bo.y + bo.h //+ 1
	if y >= pi.h {
		return false
	}
	for x := bo.x; x < (bo.x + bo.w); x++ {
		r, g, b, a := pi.At2(x, y)
		if (r != bo.r) || (g != bo.g) || (b != bo.b) || (a != bo.a) {
			return false
		}
	}
	// Expand the box 1 pixel down
	bo.h++
	return true
}

// ExpandOnce tries to expand the box to the right and downwards, once
func (pi *PixelImage) ExpandOnce(bo *Box) bool {
	if pi.ExpandRight(bo) {
		return true
	}
	return pi.ExpandDown(bo)
}

// Expand tries to expand the box to the right and downwards, until it can't expand any more.
// Returns true if the box was expanded at least once.
func (pi *PixelImage) Expand(bo *Box) (expanded bool) {
	for {
		if !pi.ExpandOnce(bo) {
			break
		}
		expanded = true
	}
	return
}

// singleHex returns a single digit hex number, as a string
// the numbers are not rounded, just floored
func singleHex(x byte) string {
	hex := strconv.FormatInt(int64(x), 16)
	if len(hex) < 2 {
		return "0" + hex // Prepend "0" if the hex representation is a single digit
	}
	return hex
}

// shortColorString returns a string representing a color on the short form "#000"
func shortColorString(r, g, b byte) string {
	return "#" + singleHex(r) + singleHex(g) + singleHex(b)
}

// CoverBox creates rectangles in the SVG image, and also marks the pixels as covered
// if pink is true, the rectangles will be pink
// if optimizeColors is true, the color strings will be shortened (and quantized)
func (pi *PixelImage) CoverBox(bo *Box, pink bool, optimizeColors bool) {
	// Draw the rectangle
	rect := pi.svgTag.AddRect(bo.x, bo.y, bo.w, bo.h)

	// Generate a fill color string
	var colorString string
	if pink {
		if optimizeColors {
			colorString = "#b38"
		} else {
			colorString = "#bb3388"
		}
	} else if optimizeColors {
		colorString = shortColorString(byte(bo.r), byte(bo.g), byte(bo.b))
	} else {
		colorString = fmt.Sprintf("#%02x%02x%02x", bo.r, bo.g, bo.b)
	}

	// Set the fill color
	rect.Fill(colorString)

	// Mark all covered pixels in the PixelImage
	for y := bo.y; y < (bo.y + bo.h); y++ {
		for x := bo.x; x < (bo.x + bo.w); x++ {
			pi.pixels[y*pi.w+x].covered = true
		}
	}
}
