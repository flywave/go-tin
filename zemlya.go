// Copyright (c) 2017-present FlyWave, Inc. All Rights Reserved.
// See License.txt for license information.

package tin

import "math"

func averageOf(d1, d2, d3, d4, noDataValue float64) float64 {
	count := 0
	sum := float64(0.0)
	lp := []float64{d1, d2, d3, d4}
	for d := range lp {
		if isNoData(lp[d], noDataValue) {
			continue
		}
		count++
		sum += lp[d]
	}

	if count > 0 {
		return sum / float64(count)
	}
	return math.NaN()
}

type ZemlyaMesh struct {
	RasterMesh
	Sample       RasterDouble
	Insert       RasterDouble
	Result       RasterDouble
	Used         RasterChar
	Token        RasterInt
	Candidates   CandidateList
	MaxError     float64
	Counter      int
	CurrentLevel int
	MaxLevel     int
}

func (z *ZemlyaMesh) scanTriangleLine(plane Plane, y int, x1, x2 float64, candidate Candidate, noDataValue float64) {
	startx := int(math.Ceil(Min(x1, x2)))
	endx := int(math.Floor(Max(x1, x2)))

	if startx > endx {
		return
	}

	z0 := plane.Eval(float64(startx), float64(y))
	dz := plane[0]

	for x := startx; x <= endx; x++ {
		if z.Used.Value(y, x) > 0 {
			var zv float64
			if z.CurrentLevel == z.MaxLevel {
				zv = z.Raster.Value(y, x)
			} else {
				zv = z.Insert.Value(y, x)
			}

			if !isNoData(zv, noDataValue) {
				diff := math.Abs(zv - z0)
				candidate.Consider(x, y, zv, diff)
			}
		}
		z0 += dz
	}
}

func (z *ZemlyaMesh) GreedyInsert(maxError float64) {
	z.MaxError = maxError
	z.Counter = 0
	w := z.Raster.Cols()
	h := z.Raster.Rows()

	if w > h {
		z.MaxLevel = int(math.Ceil(math.Log2(float64(w))))
	} else {
		z.MaxLevel = int(math.Ceil(math.Log2(float64(h))))
	}
	z.Sample = *NewRasterDouble(h, w, math.NaN())

	noDataValue := z.Raster.NoData.(float64)

	for level := z.MaxLevel - 1; level >= 1; level-- {
		step := z.MaxLevel - level
		for y := 0; y < h; y += int(math.Pow(2., float64(step))) {
			for x := 0; x < w; x += int(math.Pow(2., float64(step))) {
				if step == 1 {
					var v1 float64
					if y < h && x < w {
						v1 = z.Raster.Value(y, x)
					} else {
						v1 = math.NaN()
					}
					var v2 float64
					if y < h && x+1 < w {
						v2 = z.Raster.Value(y, x+1)
					} else {
						v2 = math.NaN()
					}
					var v3 float64
					if y+1 < h && x < w {
						v3 = z.Raster.Value(y+1, x)
					} else {
						v3 = math.NaN()
					}
					var v4 float64
					if y+1 < h && x+1 < w {
						v4 = z.Raster.Value(y+1, x+1)
					} else {
						v4 = math.NaN()
					}

					if y+1 < h && x+1 < w {
						z.Sample.SetValue(y+1, x+1, averageOf(v1, v2, v3, v4, noDataValue))
					}
				} else {
					co := int(math.Pow(2., float64(step)-1))
					d := int(math.Pow(2., float64(step)-2))

					var v1 float64
					if y+co-d < h && x+co-d < w {
						v1 = z.Raster.Value(y+co-d, x+co-d)
					} else {
						v1 = math.NaN()
					}

					var v2 float64
					if y+co-d < h && x+co+d < w {
						v2 = z.Raster.Value(y+co-d, x+co+d)
					} else {
						v2 = math.NaN()
					}

					var v3 float64
					if y+co+d < h && x+co-d < w {
						v3 = z.Raster.Value(y+co+d, x+co-d)
					} else {
						v3 = math.NaN()
					}

					var v4 float64
					if y+co+d < h && x+co+d < w {
						v4 = z.Raster.Value(y+co+d, x+co+d)
					} else {
						v4 = math.NaN()
					}

					if y+co < h && x+co < w {
						z.Sample.SetValue(y+co, x+co, averageOf(v1, v2, v3, v4, noDataValue))
					}
				}
			}
		}
	}

	z.repairPoint(0, 0)
	z.repairPoint(0, float64(h-1))
	z.repairPoint(float64(w-1), float64(h-1))
	z.repairPoint(float64(w-1), 0)

	z.Result = *NewRasterDouble(h, w, math.NaN())
	z.Used = *NewRasterChar(h, w, 0)
	z.Token = *NewRasterInt(h, w, 0)

	z.initMesh([2]float64{0, 0}, [2]float64{0, float64(h - 1)}, [2]float64{float64(w - 1), float64(h - 1)},
		[2]float64{float64(w - 1), 0})

	for level := 1; level <= z.MaxLevel; level++ {
		z.CurrentLevel = level
		z.Used = *NewRasterChar(h, w, 0)

		if level >= 5 && level <= z.MaxLevel-1 {
			step := z.MaxLevel - level

			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					zv := z.Insert.Value(y, x)
					if isNoData(zv, noDataValue) {
						continue
					}
					z.Insert.SetValue(y, x, z.Raster.Value(y, x))
				}
			}

			for y := 0; y < h; y += int(math.Pow(2., float64(step))) {
				for x := 0; x < w; x += int(math.Pow(2., float64(step))) {
					co := int(math.Pow(2., float64(step)-1))
					if y+co < h && x+co < w {
						z.Insert.SetValue(y+co, x+co, z.Raster.Value(y+co, x+co))
					}
				}
			}
		} else if level < z.MaxLevel {
			step := z.MaxLevel - level

			if step >= 3 {
				d := int(math.Pow(2., float64(step)-3))

				for y := 0; y < h; y++ {
					for x := 0; x < w; x++ {
						zv := z.Insert.Value(y, x)
						if isNoData(zv, noDataValue) {
							continue
						}

						var v1 float64
						if y-d < h && x-d < w {
							v1 = z.Sample.Value(y-d, x-d)
						} else {
							v1 = math.NaN()
						}

						var v2 float64
						if y-d < h && x+d < w {
							v2 = z.Sample.Value(y-d, x+d)
						} else {
							v2 = math.NaN()
						}

						var v3 float64
						if y+d < h && x-d < w {
							v3 = z.Sample.Value(y+d, x-d)
						} else {
							v3 = math.NaN()
						}

						var v4 float64
						if y+d < h && x+d < w {
							v4 = z.Sample.Value(y+d, x+d)
						} else {
							v4 = math.NaN()
						}

						avg := averageOf(v1, v2, v3, v4, noDataValue)
						if isNoData(avg, noDataValue) {
							continue
						}
						z.Insert.SetValue(y, x, avg)
					}
				}
			}

			for y := 0; y < h; y += int(math.Pow(2., float64(step))) {
				for x := 0; x < w; x += int(math.Pow(2., float64(step))) {
					co := int(math.Pow(2., float64(step)-1))
					if y+co < h && x+co < w {
						z.Insert.SetValue(y+co, x+co, z.Sample.Value(y+co, x+co))
					}
				}
			}
		}

		t := z.firstFace
		for {
			z.scanTriangle(t)
			t = t.GetLink()
			if t == nil {
				break
			}
		}

		for {
			if z.Candidates.Empty() {
				break
			}
			candidate := z.Candidates.GrabGreatest()

			if candidate.Importance < z.MaxError {
				continue
			}

			if z.Token.Value(candidate.Y, candidate.X) != int32(candidate.Token) {
				continue
			}

			z.Result.SetValue(candidate.Y, candidate.X, candidate.Z)
			z.Used.SetValue(candidate.Y, candidate.X, 1)

			z.insert([2]float64{float64(candidate.X), float64(candidate.Y)}, candidate.Triangle)

		}
	}
}

func (z *ZemlyaMesh) ScanTriangle(t *DelaunayTriangle) {
	var z_plane Plane
	z_plane = computePlane(z_plane, t, &z.Result)

	byy := [3][2]float64{t.point1(), t.point2(), t.point3()}

	orderTrianglePoints(byy)

	v0_x := byy[0][0]
	v0_y := byy[0][1]
	v1_x := byy[1][0]
	v1_y := byy[1][1]
	v2_x := byy[2][0]
	v2_y := byy[2][1]

	candidate := Candidate{X: 0, Y: 0, Z: 0.0, Importance: -math.MaxFloat64, Token: z.Counter, Triangle: t}
	z.Counter++
	dx2 := (v2_x - v0_x) / (v2_y - v0_y)
	noDataValue := z.Raster.NoData.(float64)

	if v1_y != v0_y {
		dx1 := (v1_x - v0_x) / (v1_y - v0_y)

		x1 := v0_x
		x2 := v0_x

		starty := int(v0_y)
		endy := int(v1_y)

		for y := starty; y < endy; y++ {
			z.scanTriangleLine(z_plane, y, x1, x2, candidate, noDataValue)
			x1 += dx1
			x2 += dx2
		}
	}

	if v2_y != v1_y {
		dx1 := (v2_x - v1_x) / (v2_y - v1_y)

		x1 := v1_x
		x2 := v0_x

		starty := int(v1_y)
		endy := int(v2_y)

		for y := starty; y <= endy; y++ {
			z.scanTriangleLine(z_plane, y, x1, x2, candidate, noDataValue)
			x1 += dx1
			x2 += dx2
		}
	}

	z.Token.SetValue(candidate.Y, candidate.X, int32(candidate.Token))
	z.Candidates.Push(candidate)
}
