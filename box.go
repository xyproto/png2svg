package png2svg

import (
	"fmt"
	"math/rand"

	"github.com/xyproto/onthefly"
)

type Box struct {
	x int
	y int
	w int
	h int
	r int
	g int
	b int
}

func (pi *PixelImage) CreateRandomBox(checkIfPossible bool) *Box {
	w := 1
	h := 1
	var x, y, r, g, b int
	for !checkIfPossible || !pi.Done() {
		// Find a random placement for (x,y), for a box of size (1,1)
		x = rand.Intn(pi.w)
		y = rand.Intn(pi.h)
		if pi.verbose {
			fmt.Printf("Random box at (%d, %d)\n", x, y)
		}
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

func (pi *PixelImage) CreateBox(x, y int) *Box {
	w := 1
	h := 1
	var r, g, b int
	if pi.Covered(x, y) {
		panic("CreateBox at location that was already covered")
	}
	r, g, b = pi.At(x, y)
	// Create a box at that placement, with width 1 and height 1
	// Return the box
	return &Box{x, y, w, h, r, g, b}
}

// Expand a box to the left, if all new pixels have the same color
func (pi *PixelImage) ExpandLeft(bo *Box) bool {
	// Loop from box top left (-1,0) to box bot left (-1,0)
	x := bo.x - 1
	if x <= 0 {
		return false
	}
	for y := bo.y; y < (bo.y + bo.h); y++ {
		r, g, b := pi.At(x, y)
		if (r != bo.r) || (g != bo.g) || (b != bo.b) {
			return false
		}
	}
	// Expand the box 1 pixel to the left
	bo.w++
	bo.x--
	return true
}

// Expand a box upwards, if all new pixels have the same color
func (pi *PixelImage) ExpandUp(bo *Box) bool {
	// Loop from box top left to box top right
	y := bo.y - 1
	if y <= 0 {
		return false
	}
	for x := bo.x; x < (bo.x + bo.w); x++ {
		r, g, b := pi.At(x, y)
		if (r != bo.r) || (g != bo.g) || (b != bo.b) {
			return false
		}
	}
	// Expand the box 1 pixel up
	bo.h++
	bo.y--
	return true
}

// Expand a box to the right, if all new pixels have the same color
func (pi *PixelImage) ExpandRight(bo *Box) bool {
	// Loop from box top right (+1,0) to box bot right (+1,0)
	x := bo.x + bo.w + 1
	if x >= pi.w {
		return false
	}
	for y := bo.y; y < (bo.y + bo.h); y++ {
		r, g, b := pi.At(x, y)
		if (r != bo.r) || (g != bo.g) || (b != bo.b) {
			return false
		}
	}
	// Expand the box 1 pixel to the right
	bo.w++
	return true
}

// Expand a box downwards, if all new pixels have the same color
func (pi *PixelImage) ExpandDown(bo *Box) bool {
	// Loop from box bot left to box bot right
	y := bo.y + bo.h + 1
	if y >= pi.h {
		return false
	}
	for x := bo.x; x < (bo.x + bo.w); x++ {
		r, g, b := pi.At(x, y)
		if (r != bo.r) || (g != bo.g) || (b != bo.b) {
			return false
		}
	}
	// Expand the box 1 pixel down
	bo.h++
	return true
}

// ExpandRandom tries to expand the box in a random directions, once
func (pi *PixelImage) ExpandRandomOnce(bo *Box) (expanded bool) {
	switch rand.Intn(4) {
	case 0:
		if pi.ExpandRight(bo) {
			return true
		}
	case 1:
		if pi.ExpandDown(bo) {
			return true
		}
	case 2:
		if pi.ExpandLeft(bo) {
			return true
		}
	case 3:
		if pi.ExpandUp(bo) {
			return true
		}
	}
	return false
}

// ExpandOnce tries to expand the box in all directions, once
func (pi *PixelImage) ExpandOnce(bo *Box) (expanded bool) {
	if pi.ExpandRight(bo) {
		return true
	}
	if pi.ExpandDown(bo) {
		return true
	}
	//if pi.ExpandLeft(bo) {
	//	return true
	//}
	//if pi.ExpandUp(bo) {
	//	return true
	//}
	return
}

// Expand tries to expand the box in all directions, until it can't expand any more.
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

// ExpandRandom tries to expand the box randomly in all directions, until it can't expand any more.
// Returns true if the box was expanded at least once.
func (pi *PixelImage) ExpandRandom(bo *Box) (expanded bool) {
	for {
		if !pi.ExpandRandomOnce(bo) {
			break
		}
		expanded = true
	}
	return
}

// CoverBox creates rectangles in the SVG image, and also marks the pixels as covered
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
