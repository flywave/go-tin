package tin

import (
	"container/heap"
	"fmt"
	"io"
	"math"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-geoid"
)

var (
	EPSG4326 = geo.NewProj(4326)
	EPSG3857 = geo.NewProj(3857)
)

type PQ []*Candidate

func (pq PQ) Len() int { return len(pq) }

func (pq PQ) Less(i, j int) bool {
	return pq[i].Importance < pq[j].Importance
}

func (pq PQ) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PQ) Push(x interface{}) {
	temp := x.(*Candidate)
	temp.index = len(*pq)
	*pq = append(*pq, temp)
	// st.Slice(*pq, func(i, j int) bool {
	// 	return (*pq)[i].Importance > (*pq)[j].Importance
	// })
}

func (pq *PQ) Pop() interface{} {
	temp := (*pq)[0]
	temp.index = -1
	l := len(*pq)
	if l > 1 {
		*pq = (*pq)[1:]
	} else {
		*pq = PQ{}
	}
	return temp
}

func (pq *PQ) Update(entry *Candidate, importance float64) {
	entry.Importance = importance
	heap.Fix(pq, entry.index)
}

type Candidate struct {
	X          int
	Y          int
	Z          float64
	Importance float64
	Token      int
	Triangle   *DelaunayTriangle
	index      int
}

func (c *Candidate) Consider(sx, sy int, sz, imp float64) {
	if imp > c.Importance {
		c.X = sx
		c.Y = sy
		c.Z = sz
		c.Importance = imp
	}
}

func (c *Candidate) Less(o Candidate) bool {
	return c.Importance < o.Importance
}

type CandidateList struct {
	Candidates PQ
}

func (cl *CandidateList) Push(candidate *Candidate) { cl.Candidates.Push(candidate) }

func (cl *CandidateList) Size() int { return cl.Candidates.Len() }

func (cl *CandidateList) Empty() bool { return cl.Candidates.Len() == 0 }

func (cl *CandidateList) GrabGreatest() *Candidate {
	if cl.Empty() {
		return &Candidate{}
	}

	candidate := cl.Candidates.Pop()
	return candidate.(*Candidate)
}

func orderTrianglePoints(p *[3][2]float64) {
	if p[0][1] > p[1][1] {
		tmp := p[0]
		p[0] = p[1]
		p[1] = tmp
	}
	if p[1][1] > p[2][1] {
		tmp := p[1]
		p[1] = p[2]
		p[2] = tmp
	}
	if p[0][1] > p[1][1] {
		tmp := p[0]
		p[0] = p[1]
		p[1] = tmp
	}
}

func isNoData(value, noDataValue float64) bool {
	return math.IsNaN(value) || value == noDataValue
}

func computePlane(t *DelaunayTriangle, raster *RasterDouble) Plane {
	p1 := t.point1()
	p2 := t.point2()
	p3 := t.point3()

	v1 := [3]float64{p1[0], p1[1], raster.Value(int(p1[1]), int(p1[0]))}
	v2 := [3]float64{p2[0], p2[1], raster.Value(int(p2[1]), int(p2[0]))}
	v3 := [3]float64{p3[0], p3[1], raster.Value(int(p3[1]), int(p3[0]))}

	return *NewPlane(v1, v2, v3)
}

type RasterMesh struct {
	DelaunayMesh
	Raster  *RasterDouble
	SrcProj geo.Proj            // 原始坐标系
	Datum   geoid.VerticalDatum // 高程基准
	Offset  float64             // 高程偏移
}

func (r *RasterMesh) LoadRaster(raster *RasterDouble) {
	r.Raster = raster
}

func (r *RasterMesh) repairPoint(px, py float64) {
	x, y := int(px), int(py)
	noDataValue := r.Raster.NoData.(float64)

	// 检查当前点是否有效
	currentVal := r.Raster.Value(y, x)
	if !isNoData(currentVal, noDataValue) {
		return // 点已有有效值，无需修复
	}

	// 采样最近的有效平均值
	z := SampleNearestValidAvg(r.Raster, y, x, 3) // 搜索半径设为3

	if isNoData(z, noDataValue) {
		r.Raster.SetValue(y, x, 0.0) // 默认值
	} else {
		r.Raster.SetValue(y, x, z)
	}
}

func (r *RasterMesh) getElevation(y, x int) float64 {
	currentVal := r.Raster.Value(y, x)
	if r.SrcProj == nil {
		return currentVal
	}
	xCoord := r.Raster.ColToX(x)
	yCoord := r.Raster.RowToY(y)
	pt, _ := transformPoint(r.SrcProj, EPSG4326, xCoord, yCoord)

	// 高程基准转换
	if r.Datum == geoid.HAE {
		return currentVal + r.Offset
	}
	g := geoid.NewGeoid(r.Datum, false)
	return g.ConvertHeight(pt[1], pt[0], currentVal, geoid.GEOIDTOELLIPSOID)
}

// 在RasterMesh结构体下方添加新方法
func (r *RasterMesh) ExportToPLY(w io.Writer) error {
	// 写入PLY头部
	header := fmt.Sprintf(`ply
format ascii 1.0
element vertex %d
property float x
property float y
property float z
end_header
`, r.Raster.Rows()*r.Raster.Cols())

	if _, err := w.Write([]byte(header)); err != nil {
		return err
	}

	// 遍历所有栅格点
	for y := 0; y < r.Raster.Rows(); y++ {
		for x := 0; x < r.Raster.Cols(); x++ {
			// 获取高程值（已包含坐标转换逻辑）
			z := r.getElevation(y, x)

			// 获取地理坐标
			xCoord := r.Raster.ColToX(x)
			yCoord := r.Raster.RowToY(y)

			// 坐标转换（与getElevation保持一致）
			pt, _ := transformPoint(r.SrcProj, EPSG3857, xCoord, yCoord)

			// 写入顶点数据
			line := fmt.Sprintf("%f %f %f\n", pt[0], pt[1], z)
			if _, err := w.Write([]byte(line)); err != nil {
				return err
			}
		}
	}
	return nil
}
