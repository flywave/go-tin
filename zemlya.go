package tin

import (
	"math"
)

// 预定义常量避免魔法数
const (
	initialToken = -1
)

// 使用切片代替可变参数提高可读性
func averageOf(noDataValue float64, values ...float64) float64 {
	sum := 0.0
	count := 0
	for _, v := range values {
		if !isNoData(v, noDataValue) {
			sum += v
			count++
		}
	}
	if count == 0 {
		return math.NaN()
	}
	return sum / float64(count)
}

// https://isprs-archives.copernicus.org/articles/XLI-B2/459/2016/isprs-archives-XLI-B2-459-2016.pdf
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
	stepPowers   []int // 缓存步长幂值
}

func NewZemlyaMesh() *ZemlyaMesh {
	mesh := &ZemlyaMesh{}
	mesh.QuadEdges = NewPool(func() interface{} { return &QuadEdge{} })
	mesh.Triangles = NewPool(func() interface{} { return &DelaunayTriangle{} })
	mesh.scanTriangle = mesh.ScanTriangle
	return mesh
}

func (z *ZemlyaMesh) scanTriangleLine(plane Plane, y int, x1, x2 float64, candidate *Candidate, noDataValue float64) {
	startx := int(math.Ceil(math.Min(x1, x2)))
	endx := int(math.Floor(math.Max(x1, x2)))

	if startx > endx {
		return
	}

	z0 := plane.Eval(float64(startx), float64(y))
	dz := plane[0]

	for x := startx; x <= endx; x++ {
		if z.Used.Value(y, x) != 0 {
			continue
		}

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
		z0 += dz
	}
}

func (z *ZemlyaMesh) GreedyInsert(maxError float64) {
	z.MaxError = maxError
	z.Counter = 0
	w := z.Raster.Cols()
	h := z.Raster.Rows()
	noDataValue := z.Raster.NoData.(float64)

	// 计算最大层级并预计算步长幂值
	z.MaxLevel = int(math.Ceil(math.Log2(float64(max(w, h)))))
	z.stepPowers = make([]int, z.MaxLevel+1)
	for i := range z.stepPowers {
		z.stepPowers[i] = 1 << uint(i) // 2^i
	}

	// 初始化样本栅格
	z.Sample = NewRasterDouble(h, w, noDataValue)
	for level := z.MaxLevel - 1; level >= 1; level-- {
		step := z.MaxLevel - level
		stepSize := z.stepPowers[step]
		co := z.stepPowers[step-1]
		d := z.stepPowers[step-2]

		for y := 0; y < h; y += stepSize {
			for x := 0; x < w; x += stepSize {
				var values [4]float64
				points := [4][2]int{
					{y + co - d, x + co - d},
					{y + co - d, x + co + d},
					{y + co + d, x + co - d},
					{y + co + d, x + co + d},
				}

				for i, p := range points {
					if p[0] >= 0 && p[0] < h && p[1] >= 0 && p[1] < w {
						values[i] = z.Raster.Value(p[0], p[1])
					} else {
						values[i] = math.NaN()
					}
				}

				if yCo := y + co; yCo < h && x+co < w {
					avg := averageOf(noDataValue, values[:]...)
					if !math.IsNaN(avg) {
						z.Sample.SetValue(yCo, x+co, avg)
					}
				}
			}
		}
	}

	// 初始化关键点
	keyPoints := [][2]float64{
		{0, 0},
		{0, float64(h - 1)},
		{float64(w - 1), float64(h - 1)},
		{float64(w - 1), 0},
	}
	for _, pt := range keyPoints {
		z.repairPoint(pt[0], pt[1])
	}

	// 初始化结果栅格
	z.Result = NewRasterDouble(h, w, noDataValue)
	z.Result.Hemlines = z.Raster.Hemlines
	z.Result.SetValue(0, 0, z.Raster.Value(0, 0))
	z.Result.SetValue(h-1, 0, z.Raster.Value(h-1, 0))
	z.Result.SetValue(h-1, w-1, z.Raster.Value(h-1, w-1))
	z.Result.SetValue(0, w-1, z.Raster.Value(0, w-1))

	z.Insert = NewRasterDouble(h, w, noDataValue)
	z.Used = NewRasterChar(h, w, 0)
	z.Token = NewRasterInt(h, w, initialToken)

	// 初始化网格
	z.initMesh(keyPoints[0], keyPoints[1], keyPoints[2], keyPoints[3])

	// 层级处理
	for level := 1; level <= z.MaxLevel; level++ {
		z.CurrentLevel = level
		z.Used.Fill(0) // 重用栅格

		step := z.MaxLevel - level
		if step > 0 {
			stepSize := z.stepPowers[step]
			d := z.stepPowers[max(0, step-3)]

			// 处理插入栅格
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					if isNoData(z.Insert.Value(y, x), noDataValue) {
						continue
					}

					if level >= 5 && level <= z.MaxLevel-1 {
						z.Insert.SetValue(y, x, z.Raster.Value(y, x))
					} else if step >= 3 {
						var sampleValues [4]float64
						samplePoints := [4][2]int{
							{y - d, x - d},
							{y - d, x + d},
							{y + d, x - d},
							{y + d, x + d},
						}

						for i, p := range samplePoints {
							if p[0] >= 0 && p[0] < h && p[1] >= 0 && p[1] < w {
								sampleValues[i] = z.Sample.Value(p[0], p[1])
							} else {
								sampleValues[i] = math.NaN()
							}
						}

						avg := averageOf(noDataValue, sampleValues[:]...)
						if !math.IsNaN(avg) {
							z.Insert.SetValue(y, x, avg)
						}
					}
				}
			}

			// 设置样本点
			for y := 0; y < h; y += stepSize {
				for x := 0; x < w; x += stepSize {
					co := z.stepPowers[step-1]
					yCo, xCo := y+co, x+co
					if yCo < h && xCo < w {
						z.Insert.SetValue(yCo, xCo, z.Sample.Value(yCo, xCo))
					}
				}
			}
		}

		// 处理候选点
		for t := z.firstFace; t != nil; t = t.GetLink() {
			z.ScanTriangle(t)
		}

		for !z.Candidates.Empty() {
			candidate := z.Candidates.GrabGreatest()
			if candidate.Importance < z.MaxError || z.Token.Value(candidate.Y, candidate.X) != int32(candidate.Token) {
				continue
			}

			z.Result.SetValue(candidate.Y, candidate.X, candidate.Z)
			z.Used.SetValue(candidate.Y, candidate.X, 1)
			z.insert([2]float64{float64(candidate.X), float64(candidate.Y)}, candidate.Triangle)
		}
	}
}

func (z *ZemlyaMesh) ScanTriangle(t *DelaunayTriangle) {
	zPlane := computePlane(t, z.Result)
	points := [3][2]float64{t.point1(), t.point2(), t.point3()}
	orderTrianglePoints(&points)

	v0, v1, v2 := points[0], points[1], points[2]
	noDataValue := z.Raster.NoData.(float64)
	candidate := &Candidate{
		Importance: -math.MaxFloat64,
		Token:      z.Counter,
		Triangle:   t,
	}
	z.Counter++

	// 扫描三角形边
	if v1[1] != v0[1] {
		dx1 := (v1[0] - v0[0]) / (v1[1] - v0[1])
		dx2 := (v2[0] - v0[0]) / (v2[1] - v0[1])
		x1, x2 := v0[0], v0[0]

		for y := int(v0[1]); y <= int(v1[1]); y++ {
			z.scanTriangleLine(zPlane, y, x1, x2, candidate, noDataValue)
			x1 += dx1
			x2 += dx2
		}
	}

	if v2[1] != v1[1] {
		dx1 := (v2[0] - v1[0]) / (v2[1] - v1[1])
		dx2 := (v2[0] - v0[0]) / (v2[1] - v0[1])
		x1, x2 := v1[0], v0[0]+dx2*(v1[1]-v0[1])

		for y := int(v1[1]); y <= int(v2[1]); y++ {
			z.scanTriangleLine(zPlane, y, x1, x2, candidate, noDataValue)
			x1 += dx1
			x2 += dx2
		}
	}

	if candidate.Importance >= z.MaxError {
		z.Token.SetValue(candidate.Y, candidate.X, int32(candidate.Token))
		z.Candidates.Push(candidate)
	}
}

func (z *ZemlyaMesh) ToMesh() *Mesh {
	w, h := z.Raster.Cols(), z.Raster.Rows()
	noDataValue := z.Raster.NoData.(float64)
	mesh := &Mesh{}
	vertexID := NewRasterInt(h, w, -1)
	vertices := make([]Vertex, 0, w*h/2)
	normals := make([]Normal, 0, w*h/2)

	// 收集有效顶点
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if zv := z.Result.Value(y, x); !isNoData(zv, noDataValue) {
				v := Vertex{z.Raster.ColToX(x), z.Raster.RowToY(y), zv}
				if z.Raster.transform != nil {
					v = z.Raster.transform(&v)
				}
				vertexID.SetValue(y, x, int32(len(vertices)))
				vertices = append(vertices, v)
				normals = append(normals, Normal{})
				mesh.updateBBox(v)
			}
		}
	}

	// 处理三角形面
	var faces []Face
	for t := z.firstFace; t != nil; t = t.GetLink() {
		p1, p2, p3 := t.point1(), t.point2(), t.point3()
		idx1 := vertexID.Value(int(p1[1]), int(p1[0]))
		idx2 := vertexID.Value(int(p2[1]), int(p2[0]))
		idx3 := vertexID.Value(int(p3[1]), int(p3[0]))

		if idx1 < 0 || idx2 < 0 || idx3 < 0 {
			continue
		}

		var face Face
		if IsCCW(p1, p2, p3) {
			face = Face{VertexIndex(idx3), VertexIndex(idx2), VertexIndex(idx1)}
		} else {
			face = Face{VertexIndex(idx1), VertexIndex(idx2), VertexIndex(idx3)}
		}

		faces = append(faces, face)
		updateNormals(vertices, normals, face)
	}

	// 归一化法向量
	for i := range normals {
		n := &normals[i]
		if length := math.Sqrt(n[0]*n[0] + n[1]*n[1] + n[2]*n[2]); length > 0 {
			n[0] /= length
			n[1] /= length
			n[2] /= length
		}
	}

	mesh.initFromDecomposed(vertices, faces, normals)
	return mesh
}

// 辅助函数
func updateNormals(vertices []Vertex, normals []Normal, face Face) {
	v0 := vertices[face[0]]
	v1 := vertices[face[1]]
	v2 := vertices[face[2]]

	edge1 := [3]float64{v1[0] - v0[0], v1[1] - v0[1], v1[2] - v0[2]}
	edge2 := [3]float64{v2[0] - v0[0], v2[1] - v0[1], v2[2] - v0[2]}

	normal := [3]float64{
		edge1[1]*edge2[2] - edge1[2]*edge2[1],
		edge1[2]*edge2[0] - edge1[0]*edge2[2],
		edge1[0]*edge2[1] - edge1[1]*edge2[0],
	}

	for _, idx := range face {
		normals[idx][0] += normal[0]
		normals[idx][1] += normal[1]
		normals[idx][2] += normal[2]
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
