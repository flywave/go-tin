package tin

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrientation(t *testing.T) {
	tests := []struct {
		name     string
		a, b, c  [2]float64
		expected float64
	}{
		{
			name:     "collinear points",
			a:        [2]float64{0, 0},
			b:        [2]float64{1, 1},
			c:        [2]float64{2, 2},
			expected: 0,
		},
		{
			name:     "counter-clockwise orientation",
			a:        [2]float64{0, 0},
			b:        [2]float64{1, 0},
			c:        [2]float64{0, 1},
			expected: 1,
		},
		{
			name:     "clockwise orientation",
			a:        [2]float64{0, 0},
			b:        [2]float64{0, 1},
			c:        [2]float64{1, 0},
			expected: -1,
		},
		{
			name:     "all points identical",
			a:        [2]float64{5, 5},
			b:        [2]float64{5, 5},
			c:        [2]float64{5, 5},
			expected: 0,
		},
		{
			name:     "a and b identical",
			a:        [2]float64{2, 2},
			b:        [2]float64{2, 2},
			c:        [2]float64{4, 4},
			expected: 0,
		},
		{
			name:     "a and c identical",
			a:        [2]float64{3, 3},
			b:        [2]float64{6, 6},
			c:        [2]float64{3, 3},
			expected: 0,
		},
		{
			name:     "vertical line collinear",
			a:        [2]float64{0, 0},
			b:        [2]float64{0, 1},
			c:        [2]float64{0, 2},
			expected: 0,
		},
		{
			name:     "horizontal line collinear",
			a:        [2]float64{0, 0},
			b:        [2]float64{1, 0},
			c:        [2]float64{2, 0},
			expected: 0,
		},
		{
			name:     "large coordinate values",
			a:        [2]float64{1000000, 1000000},
			b:        [2]float64{2000000, 2000000},
			c:        [2]float64{1000000, 2000000},
			expected: 1000000000000, // Changed from -1e12 to 1e12
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Orientation(tt.a, tt.b, tt.c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsCCW(t *testing.T) {
	tests := []struct {
		name     string
		a, b, c  [2]float64
		expected bool
	}{
		{
			name:     "counter-clockwise triangle",
			a:        [2]float64{0, 0},
			b:        [2]float64{1, 0},
			c:        [2]float64{0, 1},
			expected: true,
		},
		{
			name:     "clockwise triangle",
			a:        [2]float64{0, 0},
			b:        [2]float64{0, 1},
			c:        [2]float64{1, 0},
			expected: false,
		},
		{
			name:     "collinear points",
			a:        [2]float64{0, 0},
			b:        [2]float64{1, 1},
			c:        [2]float64{2, 2},
			expected: false,
		},
		{
			name:     "another counter-clockwise case",
			a:        [2]float64{1, 1},
			b:        [2]float64{2, 3},
			c:        [2]float64{3, 1},
			expected: false, // Changed from true to false
		},
		{
			name:     "another clockwise case",
			a:        [2]float64{1, 1},
			b:        [2]float64{3, 1},
			c:        [2]float64{2, 3},
			expected: true, // Changed from false to true
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCCW(tt.a, tt.b, tt.c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInTriangle(t *testing.T) {
	tests := []struct {
		name     string
		a, b, c  [2]float64
		p        [2]float64
		expected bool
	}{
		{
			name:     "point inside triangle",
			a:        [2]float64{0, 0},
			b:        [2]float64{2, 0},
			c:        [2]float64{0, 2},
			p:        [2]float64{0.5, 0.5},
			expected: true,
		},
		{
			name:     "point outside triangle",
			a:        [2]float64{0, 0},
			b:        [2]float64{2, 0},
			c:        [2]float64{0, 2},
			p:        [2]float64{1.5, 1.5},
			expected: false,
		},
		{
			name:     "point on vertex",
			a:        [2]float64{0, 0},
			b:        [2]float64{2, 0},
			c:        [2]float64{0, 2},
			p:        [2]float64{0, 0},
			expected: true,
		},
		{
			name:     "point on edge",
			a:        [2]float64{0, 0},
			b:        [2]float64{2, 0},
			c:        [2]float64{0, 2},
			p:        [2]float64{1, 0},
			expected: true,
		},
		{
			name:     "point outside with negative coordinates",
			a:        [2]float64{-1, -1},
			b:        [2]float64{1, -1},
			c:        [2]float64{-1, 1},
			p:        [2]float64{1, 1},
			expected: false,
		},
		{
			name:     "point inside with negative coordinates",
			a:        [2]float64{-1, -1},
			b:        [2]float64{1, -1},
			c:        [2]float64{-1, 1},
			p:        [2]float64{-0.5, -0.5},
			expected: true,
		},
		{
			name:     "degenerate triangle (colinear points)",
			a:        [2]float64{0, 0},
			b:        [2]float64{1, 1},
			c:        [2]float64{2, 2},
			p:        [2]float64{1.5, 1.5},
			expected: false,
		},
		{
			name:     "point very close to edge",
			a:        [2]float64{0, 0},
			b:        [2]float64{2, 0},
			c:        [2]float64{0, 2},
			p:        [2]float64{0.0001, 0.0001},
			expected: true,
		},
		{
			name:     "point very close but outside",
			a:        [2]float64{0, 0},
			b:        [2]float64{2, 0},
			c:        [2]float64{0, 2},
			p:        [2]float64{-0.0001, -0.0001},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InTriangle(tt.a, tt.b, tt.c, tt.p)
			assert.Equal(t, tt.expected, result, "InTriangle(%v, %v, %v, %v)", tt.a, tt.b, tt.c, tt.p)
		})
	}
}

func TestInCircumcircle(t *testing.T) {
	tests := []struct {
		name       string
		a, b, c, d [2]float64
		expected   bool
	}{
		{
			name:     "point inside circumcircle",
			a:        [2]float64{0, 0},
			b:        [2]float64{2, 0},
			c:        [2]float64{0, 2},
			d:        [2]float64{1, 1}, // center point
			expected: true,
		},
		{
			name:     "point outside circumcircle",
			a:        [2]float64{0, 0},
			b:        [2]float64{2, 0},
			c:        [2]float64{0, 2},
			d:        [2]float64{3, 3}, // far outside
			expected: false,
		},
		{
			name:     "point exactly on circumcircle",
			a:        [2]float64{0, 0},
			b:        [2]float64{2, 0},
			c:        [2]float64{0, 2},
			d:        [2]float64{2, 2}, // on the circle
			expected: false,            // within EPS tolerance
		},
		{
			name:     "collinear points (degenerate case)",
			a:        [2]float64{0, 0},
			b:        [2]float64{1, 0},
			c:        [2]float64{2, 0},
			d:        [2]float64{1.5, 0}, // on the line
			expected: false,
		},
		{
			name:     "nearly collinear points",
			a:        [2]float64{0, 0},
			b:        [2]float64{1, 0.0000001},
			c:        [2]float64{2, 0},
			d:        [2]float64{1, 0.0000002}, // slightly off the line
			expected: true,                     // inside due to curvature
		},
		{
			name:     "large coordinates",
			a:        [2]float64{1000, 1000},
			b:        [2]float64{1002, 1000},
			c:        [2]float64{1000, 1002},
			d:        [2]float64{1001, 1001}, // center
			expected: true,
		},
		{
			name:     "small coordinates",
			a:        [2]float64{0.0001, 0.0001},
			b:        [2]float64{0.0002, 0.0001},
			c:        [2]float64{0.0001, 0.0002},
			d:        [2]float64{0.00015, 0.00015}, // center
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InCircumcircle(tt.a, tt.b, tt.c, tt.d)
			if result != tt.expected {
				t.Errorf("test case '%s' failed: expected %v, got %v", tt.name, tt.expected, result)
			}
		})
	}
}

func TestCircumcenter(t *testing.T) {
	tests := []struct {
		name     string
		a, b, c  [2]float64
		expected [2]float64
		wantErr  bool
	}{
		{
			name:     "equilateral triangle",
			a:        [2]float64{0, 0},
			b:        [2]float64{2, 0},
			c:        [2]float64{1, math.Sqrt(3)},
			expected: [2]float64{1, math.Sqrt(3) / 3},
		},
		{
			name:     "right-angled triangle",
			a:        [2]float64{0, 0},
			b:        [2]float64{4, 0},
			c:        [2]float64{0, 3},
			expected: [2]float64{2, 1.5},
		},
		{
			name:     "scalene triangle",
			a:        [2]float64{1, 1},
			b:        [2]float64{5, 2},
			c:        [2]float64{3, 6},
			expected: [2]float64{23.0 / 9.0, 59.0 / 18.0}, // 修正后的正确值
		},
		{
			name:     "points with negative coordinates",
			a:        [2]float64{-1, -1},
			b:        [2]float64{-3, -4},
			c:        [2]float64{-5, -2},
			expected: [2]float64{-2.9, -1.9}, // 修正后的正确值
		},
		{
			name:    "collinear points",
			a:       [2]float64{0, 0},
			b:       [2]float64{1, 1},
			c:       [2]float64{2, 2},
			wantErr: true,
		},
		{
			name:    "two identical points",
			a:       [2]float64{0, 0},
			b:       [2]float64{0, 0},
			c:       [2]float64{1, 1},
			wantErr: true,
		},
		{
			name:    "all identical points",
			a:       [2]float64{0, 0},
			b:       [2]float64{0, 0},
			c:       [2]float64{0, 0},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Circumcenter(tt.a, tt.b, tt.c)

			if tt.wantErr {
				ba0, ba1 := tt.b[0]-tt.a[0], tt.b[1]-tt.a[1]
				ca0, ca1 := tt.c[0]-tt.a[0], tt.c[1]-tt.a[1]
				cross := ba0*ca1 - ba1*ca0
				assert.True(t, math.Abs(cross) < 1e-15, "expected collinear points")
				assert.True(t, math.IsNaN(result[0]) && math.IsNaN(result[1]),
					"expected NaN for degenerate case")
			} else {
				assert.InDelta(t, tt.expected[0], result[0], 1e-9, "x coordinate mismatch")
				assert.InDelta(t, tt.expected[1], result[1], 1e-9, "y coordinate mismatch")

				// 验证外心到三点的距离相等
				distSq := func(p1, p2 [2]float64) float64 {
					dx := p1[0] - p2[0]
					dy := p1[1] - p2[1]
					return dx*dx + dy*dy
				}

				d1 := distSq(result, tt.a)
				d2 := distSq(result, tt.b)
				d3 := distSq(result, tt.c)

				assert.InDelta(t, d1, d2, 1e-9, "not equidistant to a and b")
				assert.InDelta(t, d1, d3, 1e-9, "not equidistant to a and c")
			}
		})
	}
}
