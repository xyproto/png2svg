package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
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

func main() {
	var config Config
	app := &cli.App{
		Name:  "png2svg",
		Usage: "Convert PNG images to SVG format",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "o",
				Value:       "-",
				Usage:       "SVG output filename",
				Destination: &config.outputFilename,
			},
			&cli.BoolFlag{
				Name:        "p",
				Usage:       "use only single pixel rectangles",
				Destination: &config.singlePixelRectangles,
			},
			&cli.BoolFlag{
				Name:        "c",
				Usage:       "color expanded rectangles pink",
				Destination: &config.colorPink,
			},
			&cli.BoolFlag{
				Name:        "v",
				Usage:       "verbose",
				Destination: &config.verbose,
			},
			&cli.BoolFlag{
				Name:        "V",
				Usage:       "version",
				Destination: &config.version,
			},
			&cli.BoolFlag{
				Name:        "l",
				Usage:       "limit colors to a maximum of 4096 (#abcdef -> #ace)",
				Destination: &config.limit,
			},
			&cli.BoolFlag{
				Name:        "q",
				Usage:       "deprecated (same as -l)",
				Destination: &config.quantize,
			},
			&cli.BoolFlag{
				Name:        "z",
				Usage:       "deprecated (same as -l)",
				Destination: &config.colorOptimize,
			},
			&cli.IntFlag{
				Name:        "n",
				Value:       0,
				Usage:       "reduce the palette to N colors",
				Destination: &config.palReduction,
			},
		},
		Action: func(c *cli.Context) error {
			if c.Bool("version") {
				fmt.Println(png2svg.VersionString)
				return nil
			}

			config.limit = config.limit || config.quantize || config.colorOptimize

			if config.colorPink {
				config.singlePixelRectangles = false
			}

			if c.Args().Len() == 0 {
				return errors.New("an input PNG filename is required")
			}
			config.inputFilename = c.Args().First()

			return Run(&config)
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}
}

// Run performs the user-selected operations
func Run(c *Config) error {
	var (
		box          *png2svg.Box
		x, y         int
		expanded     bool
		lastx, lasty int
		lastLine     int // one message per line / y coordinate
		done         bool
	)

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

	percentage := 0
	lastPercentage := 0

	if !c.singlePixelRectangles {
		if c.verbose {
			fmt.Print("Placing rectangles... 0%")
		}

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

			// Use the expanded box. Color pink if it is > 1x1, and colorPink is true
			pi.CoverBox(box, expanded && c.colorPink, c.limit)

			// Check if we are done, searching from the current x,y
			done = pi.Done(x, y)
		}

		if c.verbose {
			png2svg.Erase(len(fmt.Sprintf("%d%%", lastPercentage)))
			fmt.Println("100%")
		}

	}

	if c.singlePixelRectangles {
		// Cover all remaining pixels with rectangles of size 1x1
		if c.verbose {
			percentage = 0
			lastPercentage = 0
			fmt.Print("Placing 1x1 rectangles... 0%")
			pi.CoverAllPixelsCallback(func(currentIndex, totalLength int) {
				lastPercentage = percentage
				percentage = int((float64(currentIndex) / float64(totalLength)) * 100.0)
				png2svg.Erase(len(fmt.Sprintf("%d%%", lastPercentage)))
				fmt.Printf("%d%%", percentage)

			}, 1024) // update the status for every 1024 pixels (+ at the start and end)
			png2svg.Erase(len(fmt.Sprintf("%d%%", lastPercentage)))
			fmt.Println("100%")
		} else {
			pi.CoverAllPixels()
		}
	}

	// Write the SVG image to outputFilename
	return pi.WriteSVG(c.outputFilename)
}
