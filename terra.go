// Copyright (c) 2017-present FlyWave, Inc. All Rights Reserved.
// See License.txt for license information.

package tin

import (
	"container/heap"
	"math"
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

func (pq *PQ) update(entry *Candidate, importance float64) {
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
	return value == math.NaN() || value == noDataValue
}

func computePlane(plane Plane, t *DelaunayTriangle, raster *RasterDouble) Plane {
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
	z := SampleNearestValidAvg(r.Raster, int(py), int(px), 1)
	no_data_value := r.Raster.NoData.(float64)
	if isNoData(z, no_data_value) {
		r.Raster.SetValue(int(py), int(px), 0.0)
	} else {
		r.Raster.SetValue(int(py), int(px), z)
	}
}
