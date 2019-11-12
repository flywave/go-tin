package tin

import "math"

const (
	MAX_AVERAGING_SAMPLES = 64
)

func average(toAverage [MAX_AVERAGING_SAMPLES]float64, avgCount int) float64 {
	if avgCount == 0 {
		return math.NaN()
	}
	sum := float64(0)
	for i := 0; i < MAX_AVERAGING_SAMPLES; i++ {
		sum += toAverage[i]
	}
	avg := sum / float64(avgCount)
	return avg
}

func MinInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func AverageNanArr(toAverage []float64) float64 {
	sum := float64(0)
	avgCount := 0
	for i := 0; i < len(toAverage); i++ {
		v := toAverage[i]
		if v != math.NaN() {
			sum += v
			avgCount++
		}
	}
	if avgCount == 0 {
		return math.NaN()
	}
	return sum / float64(avgCount)
}

func SafeGetPixel(src *RasterDouble, w, h, r, c int64) float64 {
	if r >= 0 && r < h && c >= 0 && c < w {
		return src.Value(int(r), int(c))
	}
	return math.NaN()
}

func SubSampleRaster3x3(src *RasterDouble, noDataValue float64, w, h, r, c int64) float64 {
	var centerPixel float64
	var crossPixels [4]float64
	var diagPixels [4]float64

	diagPixels[0] = SafeGetPixel(src, w, h, r-1, c-1) // top-left
	crossPixels[0] = SafeGetPixel(src, w, h, r-1, c)  // top
	diagPixels[1] = SafeGetPixel(src, w, h, r-1, c+1) // top-right
	crossPixels[1] = SafeGetPixel(src, w, h, r, c-1)  // center-left
	centerPixel = SafeGetPixel(src, w, h, r, c)       // center-center
	crossPixels[2] = SafeGetPixel(src, w, h, r, c+1)  // center-right
	diagPixels[2] = SafeGetPixel(src, w, h, r+1, c-1) // bottom-left
	crossPixels[3] = SafeGetPixel(src, w, h, r+1, c)  // bottom-center
	diagPixels[3] = SafeGetPixel(src, w, h, r+1, c+1) // bottom-right

	if centerPixel == noDataValue {
		centerPixel = math.NaN()
	}

	for i := 0; i < 4; i++ {
		if diagPixels[i] == noDataValue {
			diagPixels[i] = math.NaN()
		}
		if crossPixels[i] == noDataValue {
			crossPixels[i] = math.NaN()
		}
	}

	crossAvg := AverageNanArr(crossPixels[:])
	diagAvg := AverageNanArr(diagPixels[:])

	weighted := [6]float64{
		centerPixel, centerPixel, centerPixel, crossAvg, crossAvg, diagAvg,
	}
	weightedAvg := AverageNanArr(weighted[:])
	return weightedAvg
}

func SampleNearestValidAvg(src *RasterDouble, _row, _column int, minAveragingSamples int) float64 {
	minAveragingSamples = MinInt(minAveragingSamples, MAX_AVERAGING_SAMPLES)

	row := _row
	column := _column
	w := src.Cols()
	h := src.Rows()

	maxRadius := int64(math.Sqrt(float64(w*w + h*h)))
	noDataValue := src.NoData.(float64)

	z := float64(0)
	if row < h && column < w {
		z = src.Value(row, column)
	}
	if !isNoData(z, noDataValue) {
		return z
	}

	var toAverage [MAX_AVERAGING_SAMPLES]float64
	for i := range toAverage {
		toAverage[i] = 0.0
	}

	avgCount := 0

	putpixel := func(x, y int64) {
		destR := int64(row) + y
		destC := int64(column) + x
		z := SubSampleRaster3x3(src, noDataValue, int64(w), int64(h), destR, destC)
		if !isNoData(z, noDataValue) {
			toAverage[avgCount] = z
			avgCount++
		}
	}

	for radius := int64(2); radius <= maxRadius && avgCount < minAveragingSamples; radius++ {
		x := radius - 1
		y := int64(0)
		dx := int64(1)
		dy := int64(1)
		err := int64(dx) - (radius / 2)

		for {
			if x >= y {
				break
			}
			putpixel(x, y)
			putpixel(y, x)
			putpixel(-y, x)
			putpixel(-x, y)
			putpixel(-x, -y)
			putpixel(-y, -x)
			putpixel(y, -x)
			putpixel(x, -y)

			if err <= 0 {
				y++
				err += dy
				dy += 2
			} else {
				x--
				dx += 2
				err += dx - (radius / 2)
			}
		}
	}

	if avgCount == 1 {
		return toAverage[0]
	}

	return average(toAverage, avgCount)
}
