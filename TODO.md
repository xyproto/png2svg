# TODO

- [ ] See if randomly placed expanding rectangles gives better results.
- [ ] See if bringing expanding circles into the mix gives better results (with a separate cover bool for circles?).
- [ ] See if bringing expanding polygons into the mix gives better results (with a separate cover bool for polygons?).
- [ ] Don't check `r`, `g` and then `b`. Even though it is short-circuited, try using a single value for checking if the color matches.
- [ ] Remove panics that were used during debugging, but are not needed anymore.
- [ ] Test with partially transparent PNG images.
