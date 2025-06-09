package tin

import (
	"math"
)

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
	Sample       *RasterDouble
	Insert       *RasterDouble
	Result       *RasterDouble
	Used         *RasterChar
	Token        *RasterInt
	Candidates   CandidateList
	MaxError     float64
	Counter      int
	CurrentLevel int
	MaxLevel     int
}

func NewZemlyaMesh() *ZemlyaMesh {
	mesh := &ZemlyaMesh{}
	mesh.QuadEdges = NewPool()
	mesh.Triangles = NewPool()
	mesh.scanTriangle = mesh.ScanTriangle
	return mesh
}

func (z *ZemlyaMesh) scanTriangleLine(plane Plane, y int, x1, x2 float64, candidate *Candidate, noDataValue float64) {
	startx := int(math.Ceil(Min(x1, x2)))
	endx := int(math.Floor(Max(x1, x2)))

	if startx > endx {
		return
	}

	z0 := plane.Eval(float64(startx), float64(y))
	dz := plane[0]

	for x := startx; x <= endx; x++ {
		if z.Used.Value(y, x) == 0 {
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
	noDataValue := z.Raster.NoData.(float64)

	if w > h {
		z.MaxLevel = int(math.Ceil(math.Log2(float64(w))))
	} else {
		z.MaxLevel = int(math.Ceil(math.Log2(float64(h))))
	}
	z.Sample = NewRasterDouble(h, w, noDataValue)

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

	z.Result = NewRasterDouble(h, w, noDataValue)
	z.Result.Hemlines = z.Raster.Hemlines
	z.Result.SetValue(0, 0, z.Raster.Value(0, 0))
	z.Result.SetValue(h-1, 0, z.Raster.Value(h-1, 0))
	z.Result.SetValue(h-1, w-1, z.Raster.Value(h-1, w-1))
	z.Result.SetValue(0, w-1, z.Raster.Value(0, w-1))

	z.Insert = NewRasterDouble(h, w, noDataValue)

	z.Used = NewRasterChar(h, w, 0)
	z.Token = NewRasterInt(h, w, 0)

	z.initMesh([2]float64{0, 0}, [2]float64{0, float64(h - 1)}, [2]float64{float64(w - 1), float64(h - 1)},
		[2]float64{float64(w - 1), 0})

	for level := 1; level <= z.MaxLevel; level++ {
		z.CurrentLevel = level
		z.Used = NewRasterChar(h, w, 0)

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
			z.ScanTriangle(t)
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
	var zPlane Plane
	zPlane = computePlane(zPlane, t, z.Result)

	byy := [3][2]float64{t.point1(), t.point2(), t.point3()}

	orderTrianglePoints(&byy)

	v0X := byy[0][0]
	v0Y := byy[0][1]
	v1X := byy[1][0]
	v1Y := byy[1][1]
	v2X := byy[2][0]
	v2Y := byy[2][1]

	candidate := &Candidate{X: 0, Y: 0, Z: 0.0, Importance: -math.MaxFloat64, Token: z.Counter, Triangle: t}
	z.Counter++
	dx2 := (v2X - v0X) / (v2Y - v0Y)
	noDataValue := z.Raster.NoData.(float64)

	if v1Y != v0Y {
		dx1 := (v1X - v0X) / (v1Y - v0Y)

		x1 := v0X
		x2 := v0X

		starty := int(v0Y)
		endy := int(v1Y)

		for y := starty; y <= endy; y++ {
			z.scanTriangleLine(zPlane, y, x1, x2, candidate, noDataValue)
			x1 += dx1
			x2 += dx2
		}
	}

	if v2Y != v1Y {
		dx1 := (v2X - v1X) / (v2Y - v1Y)

		x1 := v1X
		x2 := v0X

		starty := int(v1Y)
		endy := int(v2Y)

		for y := starty; y <= endy; y++ {
			z.scanTriangleLine(zPlane, y, x1, x2, candidate, noDataValue)
			x1 += dx1
			x2 += dx2
		}
	}

	z.Token.SetValue(candidate.Y, candidate.X, int32(candidate.Token))
	if candidate.Importance >= z.MaxError {
		z.Candidates.Push(candidate)
	}
}

func (z *ZemlyaMesh) ToMesh() *Mesh {
	w := z.Raster.Cols()
	h := z.Raster.Rows()

	var mvertices []Vertex

	vertexID := NewRasterInt(h, w, 0)
	noDataValue := z.Raster.NoData.(float64)

	index := 0
	minx := math.MaxFloat64
	miny := math.MaxFloat64
	minz := math.MaxFloat64

	maxx := -math.MaxFloat64
	maxy := -math.MaxFloat64
	maxz := -math.MaxFloat64

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			zv := z.Result.Value(y, x)
			if !isNoData(zv, noDataValue) {
				v := Vertex{z.Raster.col2x(x), z.Raster.row2y(y), zv}
				if z.Raster.transform != nil {
					v = z.Raster.transform(&v)
				}
				minx = math.Min(minx, v[0])
				miny = math.Min(miny, v[1])
				minz = math.Min(minz, v[2])

				maxx = math.Max(maxx, v[0])
				maxy = math.Max(maxy, v[1])
				maxz = math.Max(maxz, v[2])

				mvertices = append(mvertices, v)
				vertexID.SetValue(y, x, int32(index))
				index++
			}
		}
	}

	normals := make([]Normal, len(mvertices))

	var mfaces []Face
	t := z.firstFace
	for {
		var f Face

		p1 := t.point1()
		p2 := t.point2()
		p3 := t.point3()

		if !IsCCW(p1, p2, p3) {
			f[0] = VertexIndex(vertexID.Value(int(p1[1]), int(p1[0])))
			f[1] = VertexIndex(vertexID.Value(int(p2[1]), int(p2[0])))
			f[2] = VertexIndex(vertexID.Value(int(p3[1]), int(p3[0])))
		} else {
			f[0] = VertexIndex(vertexID.Value(int(p3[1]), int(p3[0])))
			f[1] = VertexIndex(vertexID.Value(int(p2[1]), int(p2[0])))
			f[2] = VertexIndex(vertexID.Value(int(p1[1]), int(p1[0])))
		}

		mfaces = append(mfaces, f)

		v0 := mvertices[f[0]]
		v1 := mvertices[f[1]]
		v2 := mvertices[f[2]]

		edge1 := [3]float64{v1[0] - v0[0], v1[1] - v0[1], v1[2] - v0[2]}
		edge2 := [3]float64{v2[0] - v0[0], v2[1] - v0[1], v2[2] - v0[2]}

		normal := [3]float64{
			edge1[1]*edge2[2] - edge1[2]*edge2[1],
			edge1[2]*edge2[0] - edge1[0]*edge2[2],
			edge1[0]*edge2[1] - edge1[1]*edge2[0],
		}

		length := math.Sqrt(normal[0]*normal[0] + normal[1]*normal[1] + normal[2]*normal[2])
		if length > 0 {
			normal[0] /= length
			normal[1] /= length
			normal[2] /= length
		}

		normals[f[0]][0] += normal[0]
		normals[f[0]][1] += normal[1]
		normals[f[0]][2] += normal[2]

		normals[f[1]][0] += normal[0]
		normals[f[1]][1] += normal[1]
		normals[f[1]][2] += normal[2]

		normals[f[2]][0] += normal[0]
		normals[f[2]][1] += normal[1]
		normals[f[2]][2] += normal[2]

		t = t.GetLink()
		if t == nil {
			break
		}
	}

	for i := range normals {
		n := &normals[i]
		length := math.Sqrt(n[0]*n[0] + n[1]*n[1] + n[2]*n[2])
		if length > 0 {
			n[0] /= length
			n[1] /= length
			n[2] /= length
		}
	}

	mesh := new(Mesh)
	mesh.BBox[0] = [3]float64{minx, miny, minz}
	mesh.BBox[1] = [3]float64{maxx, maxy, maxz}
	mesh.initFromDecomposed(mvertices, mfaces, normals)
	return mesh
}
