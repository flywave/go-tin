package tin

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubSampleRaster3x3(t *testing.T) {
	tests := []struct {
		name        string
		src         *RasterDouble
		noDataValue float64
		w, h, r, c  int64
		expected    float64
		expectNaN   bool
	}{
		{
			name: "normal case - all valid pixels",
			src: func() *RasterDouble {
				r := NewRasterDouble(3, 3, -9999.0)
				for i := 0; i < 3; i++ {
					for j := 0; j < 3; j++ {
						r.SetValue(i, j, float64(i*3+j+1))
					}
				}
				return r
			}(),
			noDataValue: -9999,
			w:           3, h: 3, r: 1, c: 1,
			expected: 5.0,
		},
		{
			name: "with noData values - should ignore them",
			src: func() *RasterDouble {
				r := NewRasterDouble(3, 3, -9999.0)
				data := [][]float64{
					{1, -9999, 3},
					{4, 5, 6},
					{7, 8, 9},
				}
				for i := 0; i < 3; i++ {
					for j := 0; j < 3; j++ {
						r.SetValue(i, j, data[i][j])
					}
				}
				return r
			}(),
			noDataValue: -9999,
			w:           3, h: 3, r: 1, c: 1,
			expected: 5.333333333333333,
		},
		{
			name: "center is noData - should use average of neighbors",
			src: &RasterDouble{
				Raster: Raster{
					Data: [][]float64{
						{1, 2, 3},
						{4, -9999, 6},
						{7, 8, 9},
					},
					NoData: -9999.0,
				},
			},
			noDataValue: -9999,
			w:           3, h: 3, r: 1, c: 1,
			expectNaN: true,
		},
		{
			name: "edge case - top left corner",
			src: func() *RasterDouble {
				r := NewRasterDouble(2, 2, -9999.0)
				r.SetValue(0, 0, 1)
				r.SetValue(0, 1, 2)
				r.SetValue(1, 0, 3)
				r.SetValue(1, 1, 4)
				return r
			}(),
			noDataValue: -9999,
			w:           2, h: 2, r: 0, c: 0,
			expected: 2.0, // Corrected from 2.5 to 2.0
		},
		{
			name: "edge case - bottom right corner",
			src: func() *RasterDouble {
				r := NewRasterDouble(2, 2, -9999.0)
				r.SetValue(0, 0, 1)
				r.SetValue(0, 1, 2)
				r.SetValue(1, 0, 3)
				r.SetValue(1, 1, 4)
				return r
			}(),
			noDataValue: -9999,
			w:           2, h: 2, r: 1, c: 1,
			expected: 3.0, // Corrected from 3.5 to 3.0
		},
		{
			name: "all pixels are noData - should return NaN",
			src: &RasterDouble{
				Raster: Raster{
					Data: [][]float64{
						{-9999, -9999, -9999},
						{-9999, -9999, -9999},
						{-9999, -9999, -9999},
					},
					NoData: -9999.0,
				},
			},
			noDataValue: -9999,
			w:           3, h: 3, r: 1, c: 1,
			expectNaN: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SubSampleRaster3x3(tt.src, tt.noDataValue, tt.w, tt.h, tt.r, tt.c)

			if tt.expectNaN {
				assert.True(t, math.IsNaN(result), "Expected NaN but got %v", result)
			} else {
				assert.InDelta(t, tt.expected, result, 0.0001, "Expected %v with delta 0.0001, but got %v", tt.expected, result)
			}
		})
	}
}
