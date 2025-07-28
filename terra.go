package tin

import (
	"container/heap"
	"math"
)

// PQ 优先级队列实现（最大堆）
type PQ []*Candidate

func (pq PQ) Len() int { return len(pq) }

func (pq PQ) Less(i, j int) bool {
	// 修改为最大堆：重要性大的元素优先
	return pq[i].Importance > pq[j].Importance
}

func (pq PQ) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PQ) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Candidate)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PQ) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1
	*pq = old[0 : n-1]
	return item
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

func (cl *CandidateList) Push(candidate *Candidate) {
	heap.Push(&cl.Candidates, candidate) // 使用heap.Push确保堆性质
}

func (cl *CandidateList) Size() int { return cl.Candidates.Len() }

func (cl *CandidateList) Empty() bool { return cl.Candidates.Len() == 0 }

func (cl *CandidateList) GrabGreatest() *Candidate {
	if cl.Empty() {
		return nil // Return nil instead of empty Candidate
	}
	return heap.Pop(&cl.Candidates).(*Candidate)
}

func orderTrianglePoints(p *[3][2]float64) {
	if p[0][1] > p[1][1] {
		p[0], p[1] = p[1], p[0]
	}
	if p[1][1] > p[2][1] {
		p[1], p[2] = p[2], p[1]
	}
	if p[0][1] > p[1][1] {
		p[0], p[1] = p[1], p[0]
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
	Raster *RasterDouble
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
