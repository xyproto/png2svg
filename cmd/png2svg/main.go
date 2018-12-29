package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/xyproto/png2svg"
)

func init() {
	// Seed the random number generator
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {
	var (
		outputFilename        string
		colorPink             bool
		singlePixelRectangles bool
		verbose               bool
		version               bool
		quantize              bool
	)

	// TODO: Use a proper package for flag handling
	flag.StringVar(&outputFilename, "o", "-", "output SVG filename")
	flag.BoolVar(&singlePixelRectangles, "p", false, "use only single pixel rectangles")
	flag.BoolVar(&colorPink, "c", false, "color expanded rectangles pink")
	flag.BoolVar(&verbose, "v", false, "verbose")
	flag.BoolVar(&version, "V", false, "version")
	flag.BoolVar(&quantize, "q", false, "quantize colors (max 4096 colors)")

	flag.Parse()

	if version {
		fmt.Println("png2svg 1.0")
		os.Exit(0)
	}

	if colorPink {
		singlePixelRectangles = false
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "An input PNG filename is required.\n")
		os.Exit(1)
	}

	inputFilename := args[0]

	img, err := png2svg.ReadPNG(inputFilename, verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	var (
		height       = img.Bounds().Max.Y - img.Bounds().Min.Y
		pi           = png2svg.NewPixelImage(img, verbose)
		box          *png2svg.Box
		x, y         int
		expanded     bool
		lastx, lasty int
		lastLine     int // one message per line / y coordinate
		done         bool
	)

	if verbose {
		fmt.Print("Placing rectangles... ")
	}

	// Cover pixels by creating expanding rectangles, as long as there are uncovered pixels
	for !singlePixelRectangles && !done {

		// Select the first uncovered pixel, searching from the given coordinate
		x, y = pi.FirstUncovered(lastx, lasty)

		if verbose && y != lastLine {
			percentage := int((float64(y) / float64(height)) * 100.0)
			fmt.Printf("%d%% ", percentage)
			lastLine = y
		}

		// Create a box at that location
		box = pi.CreateBox(x, y)
		// Expand the box to the right and downwards, until it can not expand anymore
		expanded = pi.Expand(box)

		// NOTE: Random boxes gave worse results, even though they are expanding in all directions
		// Create a random box
		//box := pi.CreateRandomBox(false)
		// Expand the box in all directions, until it can not expand anymore
		//expanded = pi.ExpandRandom(box)

		// Use the expanded box. Color pink if it is > 1x1, and colorPink is true
		pi.CoverBox(box, expanded && colorPink, quantize)

		// Check if we are done, searching from the current x,y
		done = pi.Done(x, y)
	}
	fmt.Println("100%\nDone.")

	if singlePixelRectangles {
		// Cover all remaining pixels with rectangles of size 1x1
		pi.CoverAllPixels()
	}

	// Write the SVG image to outputFilename
	err = pi.WriteSVG(outputFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
