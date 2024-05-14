package png2svg

// shortenColorLossy optimizes hexadecimal color strings in a lossy way
func shortenColorLossy(hexColorBytes []byte) []byte {
	if len(hexColorBytes) != 7 { // Only accept colors in the format #aabbcc
		return hexColorBytes
	}
	// Just keep the first digits of each 2 digit hex number
	return []byte{'#', hexColorBytes[1], hexColorBytes[3], hexColorBytes[5]}
}

// shortenColorLossless optimizes hexadecimal color strings in a lossless way
func shortenColorLossless(hexColorBytes []byte) []byte {
	if len(hexColorBytes) != 7 { // Only accept colors in the format #aabbcc
		return hexColorBytes
	}
	// Check for lossless compression
	if hexColorBytes[1] == hexColorBytes[2] && hexColorBytes[3] == hexColorBytes[4] && hexColorBytes[5] == hexColorBytes[6] {
		return []byte{'#', hexColorBytes[1], hexColorBytes[3], hexColorBytes[5]}
	}
	// Return the original color
	return hexColorBytes
}
