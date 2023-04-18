package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/xyproto/palgen"
	"github.com/xyproto/png2svg"
)

// Config contains the results of parsing the flags and arguments
type Config struct {
	inputFilename         string
	outputFilename        string
	colorOptimize         bool
	colorPink             bool
	limit                 bool
	quantize              bool
	singlePixelRectangles bool
	verbose               bool
	version               bool
	palReduction          int
}

// NewConfigFromFlags returns a Config struct, a quit message (for -v) and/or an error
func NewConfigFromFlags() (*Config, string, error) {
	var c Config

	flag.StringVar(&c.outputFilename, "o", "-", "SVG output filename")
	flag.BoolVar(&c.singlePixelRectangles, "p", false, "use only single pixel rectangles")
	flag.BoolVar(&c.colorPink, "c", false, "color expanded rectangles pink")
	flag.BoolVar(&c.verbose, "v", false, "verbose")
	flag.BoolVar(&c.version, "V", false, "version")
	flag.BoolVar(&c.limit, "l", false, "limit colors to a maximum of 4096 (#abcdef -> #ace)")
	flag.BoolVar(&c.quantize, "q", false, "deprecated (same as -l)")
	flag.BoolVar(&c.colorOptimize, "z", false, "deprecated (same as -l)")
	flag.IntVar(&c.palReduction, "n", 0, "reduce the palette to N colors")

	flag.Parse()

	if c.version {
		return nil, png2svg.VersionString, nil
	}

	c.limit = c.limit || c.quantize || c.colorOptimize

	if c.colorPink {
		c.singlePixelRectangles = false
	}

	args := flag.Args()
	if len(args) == 0 {
		return nil, "", errors.New("an input PNG filename is required")

	}
	c.inputFilename = args[0]
	return &c, "", nil
}

// Run performs the user-selected operations
func Run() error {
	var (
		box          *png2svg.Box
		x, y         int
		expanded     bool
		lastx, lasty int
		lastLine     int // one message per line / y coordinate
		done         bool
	)

	c, quitMessage, err := NewConfigFromFlags()
	if err != nil {
		return err
	} else if quitMessage != "" {
		fmt.Println(quitMessage)
		return nil
	}

	img, err := png2svg.ReadPNG(c.inputFilename, c.verbose)
	if err != nil {
		return err
	}

	if c.palReduction > 0 {
		img, err = palgen.Reduce(img, c.palReduction)
		if err != nil {
			return fmt.Errorf("could not reduce the palette of the given image to a maximum of %d colors", c.palReduction)
		}
	}

	height := img.Bounds().Max.Y - img.Bounds().Min.Y

	pi := png2svg.NewPixelImage(img, c.verbose)
	pi.SetColorOptimize(c.limit)

	if c.verbose {
		fmt.Print("Placing rectangles... 0%")
	}

	percentage := 0
	lastPercentage := 0

	// Cover pixels by creating expanding rectangles, as long as there are uncovered pixels
	for !c.singlePixelRectangles && !done {

		// Select the first uncovered pixel, searching from the given coordinate
		x, y = pi.FirstUncovered(lastx, lasty)

		if c.verbose && y != lastLine {
			lastPercentage = percentage
			percentage = int((float64(y) / float64(height)) * 100.0)
			png2svg.Erase(len(fmt.Sprintf("%d%%", lastPercentage)))
			fmt.Printf("%d%%", percentage)
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
		pi.CoverBox(box, expanded && c.colorPink, c.limit)

		// Check if we are done, searching from the current x,y
		done = pi.Done(x, y)
	}

	if c.verbose {
		png2svg.Erase(len(fmt.Sprintf("%d%%", lastPercentage)))
		fmt.Println("100%")
	}

	if c.singlePixelRectangles {
		// Cover all remaining pixels with rectangles of size 1x1
		pi.CoverAllPixels()
	}

	// Write the SVG image to outputFilename
	return pi.WriteSVG(c.outputFilename)
}

func main() {
	if err := Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", strings.Title(err.Error()))
		os.Exit(1)
	}
}
