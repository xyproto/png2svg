package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/xyproto/onthefly"
)

func readPNG(filename string) (image.Image, error) {
	fmt.Printf("Reading %s...", filename)
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

func writeSVG(img image.Image, filename string) error {
	page, svgTag := onthefly.NewTinySVGPixels(img.Bounds().Max.X-img.Bounds().Min.X, img.Bounds().Max.Y-img.Bounds().Min.Y)
	var c color.NRGBA
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		fmt.Print(".")
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			c = color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			svgTag.Pixel(x, y, int(c.R), int(c.G), int(c.B))
		}
	}
	fmt.Println("ok")

	fmt.Printf("Writing %s...", filename)
	f2, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f2.Close()
	f2.WriteString(page.String())
	fmt.Println("ok")
	return nil
}

func main() {
	var outputFilename string
	flag.StringVar(&outputFilename, "o", "output.svg", "output SVG filename")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("An input PNG filename is required.")
		fmt.Println("Try: ./png2svg -o output.svg input.png")
		os.Exit(1)
	}

	inputFilename := args[0]
	img, err := readPNG(inputFilename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if writeSVG(img, outputFilename) != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
