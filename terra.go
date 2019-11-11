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
}

func (pq *PQ) Pop() interface{} {
	temp := (*pq)[len(*pq)-1]
	temp.index = -1
	*pq = (*pq)[0 : len(*pq)-1]
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
	Triangle   *Triangle
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

func (cl *CandidateList) Push(candidate Candidate) { cl.Candidates.Push(candidate) }

func (cl *CandidateList) Size() int { return cl.Candidates.Len() }

func (cl *CandidateList) Empty() bool { return cl.Candidates.Len() == 0 }

func (cl *CandidateList) GrabGreatest() Candidate {
	if cl.Empty() {
		return Candidate{}
	}

	candidate := cl.Candidates.Pop()
	return candidate.(Candidate)
}

func swapP(p1, p2 []float64) {
	p3 := [2]float64{p1[0], p1[1]}

	p1[0] = p2[0]
	p1[1] = p2[1]

	p2[0] = p3[0]
	p2[1] = p3[1]
}

func orderTrianglePoints(p [3][2]float64) {
	if p[0][1] > p[1][1] {
		swapP(p[0][:], p[1][:])
	}
	if p[1][1] > p[2][1] {
		swapP(p[1][:], p[2][:])
	}
	if p[0][1] > p[1][1] {
		swapP(p[0][:], p[1][:])
	}
}

func isNoData(value, noDataValue float64) bool {
	return value == math.NaN() || value == noDataValue
}

type RasterMesh struct {
	DelaunayMesh
	Raster RasterDouble
}

func (r *RasterMesh) repairPoint(px, py float64) {
	z := sampleNearestValidAvg(&r.Raster, int(py), int(px), 1)
	no_data_value := r.Raster.NoData.(float64)
	if isNoData(z, no_data_value) {
		r.Raster.SetValue(int(py), int(px), 0.0)
	} else {
		r.Raster.SetValue(int(py), int(px), z)
	}
}
