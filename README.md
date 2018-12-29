# PNG2SVG [![GoDoc](https://godoc.org/github.com/xyproto/png2svg?status.svg)](http://godoc.org/github.com/xyproto/png2svg) [![Go Report Card](https://goreportcard.com/badge/github.com/xyproto/png2svg)](https://goreportcard.com/report/github.com/xyproto/png2svg)

Go module and small utility for converting PNG files to SVG Tiny 1.2

## Features and limitations

* Draws rectangles for each region in the PNG image that can be covered by a rectangle.
* The remaining pixels are drawn with a rectangle for each pixel.
* This is not an efficient representation of PNG images!
* The conversion may be useful if you have a small PNG image or icons at sizes around 16x16, and wish to scale it up and print it out without artifacts.
* The utility is fast for small images, but larger images will take an unreasonable amount of time to convert, creating SVG files many megabytes in size.
* The resulting SVG images can be opened directly in a browser like Chromium.
* Written in pure Go, with no runtime dependencies on any external library or utility.
* Handles transparent PNG images by not drawing SVG elements for the transparent regions.
* For creating SVG images that draws a rectangle for each and every pixel, use the `-p` flag.

## Comparison

| 64x64 PNG image      | 64x64 SVG image (one rectangle per pixel) | 64x64 SVG image (optimized) | 64x64 SVG image (4096 colors) |
| -------------------- | ----------------------------------------- | --------------------------- | ----------------------------- |
| 2k                   | 236k                                      | 72k                         | 68k                           |
| ![png](img/acme.png) | ![png](img/acme_singlepixel.svg)          | ![png](img/acme.svg)        | ![png](img/acme4096.svg)      |

The Glenda bunny is from [9p.io](https://9p.io/plan9/glenda.html).

| 302x240 PNG image          | 302x240 SVG image (4096 colors)  |
| -------------------------- | -------------------------------- |
| 17k                        | 312k                             |
| ![png](img/rainforest.png) | ![png](img/rainforest4096.svg) |

The rainforest image is from [wikipedia](https://en.wikipedia.org/wiki/Landscape).

## Installation

Development version:

    go get -u github.com/xyproto/png2svg/cmd/png2svg

## Example usage

Generate an SVG image with one rectangle per pixel:

    png2svg -p -o output.svg input.png

Generate an SVG image with as few rectangles as possible (optimized):

    png2svg -o output.svg input.png

Generate an SVG image with as few rectangles as possible (4096 colors):

    png2svg -q -o output.svg input.png

## General information

* Version: 1.1.0
* Author: Alexander F. Rødseth &lt;xyproto@archlinux.org&gt;
* License: MIT
