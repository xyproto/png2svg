package png2svg

import (
	"reflect"
	"testing"
)

func TestShortenColor(t *testing.T) {
	tests := []struct {
		name           string
		hexColorBytes  []byte
		colorOptimize  bool
		expectedOutput []byte
	}{
		{
			name:           "No shorten #0c0000 with colorOptimize false",
			hexColorBytes:  []byte("#0c0000"),
			colorOptimize:  false,
			expectedOutput: []byte("#0c0000"),
		},
		{
			name:           "Lossy shorten #0c0000 with colorOptimize true",
			hexColorBytes:  []byte("#0c0000"),
			colorOptimize:  true,
			expectedOutput: []byte("#000"), // Rounded to nearest single hex digit equivalent
		},
		{
			name:           "Lossless shorten #ffffff with colorOptimize true",
			hexColorBytes:  []byte("#ffffff"),
			colorOptimize:  true,
			expectedOutput: []byte("#fff"), // Lossless compression as each pair is identical
		},
		{
			name:           "No shorten #ffffff with colorOptimize false",
			hexColorBytes:  []byte("#ffffff"),
			colorOptimize:  false,
			expectedOutput: []byte("#fff"), // Should still shorten losslessly
		},
		{
			name:           "Lossy shorten #112233 with colorOptimize true",
			hexColorBytes:  []byte("#112233"),
			colorOptimize:  true,
			expectedOutput: []byte("#123"), // Each pair different, so simplified to nearest single hex equivalent
		},
		{
			name:           "No shorten #123456 with colorOptimize false",
			hexColorBytes:  []byte("#123456"),
			colorOptimize:  false,
			expectedOutput: []byte("#123456"), // No simplification as no pairs are identical
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result []byte
			if tt.colorOptimize {
				result = shortenColorLossy(tt.hexColorBytes)
			} else {
				result = shortenColorLossless(tt.hexColorBytes)
			}
			if !reflect.DeepEqual(result, tt.expectedOutput) {
				t.Errorf("shortenColor(%s, %v) = %s, want %s", tt.hexColorBytes, tt.colorOptimize, result, tt.expectedOutput)
			}
		})
	}
}
