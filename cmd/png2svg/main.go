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
		noOptimize            bool
	)

	// TODO: Use a proper package for flag handling
	flag.StringVar(&outputFilename, "o", "-", "output SVG filename")
	flag.BoolVar(&singlePixelRectangles, "p", false, "use only single pixel rectangles")
	flag.BoolVar(&colorPink, "c", false, "color expanded rectangles pink")
	flag.BoolVar(&noOptimize, "n", false, "don't optimize the resulting SVG document")
	flag.BoolVar(&verbose, "v", false, "verbose")
	flag.BoolVar(&version, "V", false, "version")

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
		pi                  = png2svg.NewPixelImage(img, verbose)
		box                 *png2svg.Box
		x, y                int
		expandedAtLeastOnce bool
	)

	// Cover pixels by creating expanding rectangles, as long as there are uncovered pixels
	for !singlePixelRectangles && !pi.Done() {

		// Select the first uncovered pixel
		x, y = pi.FirstUncovered()

		// Create a box at that location
		box = pi.CreateBox(x, y)

		// Expand the box in all directions, until it can not expand anymore
		expandedAtLeastOnce = pi.Expand(box)

		// Use the expanded box. Color pink if it is > 1x1, and colorPink is true
		pi.CoverBox(box, expandedAtLeastOnce && colorPink)
	}

	if singlePixelRectangles {
		// Cover all remaining pixels with rectangles of size 1x1
		pi.CoverAllPixels()
	}

	// Write the SVG image to outputFilename
	err = pi.WriteSVG(outputFilename, !noOptimize)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
