# png2svg

Small utility for converting PNG files to SVG Tiny 1.2

## Features and limitations

* Draws a small filled box per pixel in the PNG.
* This is horribly inefficient for large PNG files.
* May be useful if you have a small PNG image or icons at sizes around 16x16, and wish to scale it up and print it out without artifacts.
* The utility is fast for small images, but larger images will take an unreasonable amount of time to convert, creating a SVG fil that is many megabytes large.
* The resulting SVG images can be opened directly in a browser like Chromium.
* Written in pure Go, with no runtime dependencies on any external library or utility.

## General information

* Version: 0.1
* Author: Alexander F. Rødseth &lt;xyproto@archlinux.org&gt;
* License: MIT
