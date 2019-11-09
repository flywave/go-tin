// Copyright (c) 2017-present FlyWave, Inc. All Rights Reserved.
// See License.txt for license information.

package tin

import "math"

const (
	EPS = float64(0.000000001)
)

type Vertex [3]float64
type Triangle [3]Vertex

type VertexIndex int
type Face [3]VertexIndex

type Edge struct {
	First  VertexIndex
	Second VertexIndex
}

func (e *Edge) Assign(a, b VertexIndex) {
	if a < b {
		e.First = a
		e.First = b
	} else {
		e.First = b
		e.First = a
	}
}

func (e *Edge) SharesPoint(o Edge) bool {
	return e.First == o.First || e.First == o.Second ||
		e.Second == o.First || e.Second == o.Second
}

func (e *Edge) Intersects2d(o Edge, vertices []Vertex) bool {
	p0 := vertices[e.First][0:2]
	p1 := vertices[e.Second][0:2]

	l0 := vertices[o.First][0:2]
	l1 := vertices[o.Second][0:2]

	e1_bbox := BBox2d{p0[0], p0[1], p1[0], p1[1]}
	e2_bbox := BBox2d{l0[0], l0[1], l1[0], l1[1]}

	if !e1_bbox.Intersects(e2_bbox, EPS) {
		return false
	}

	intersection_point :=
		intersect2DLinesegmentByLine(p0, p1, l0, l1)
	if (intersection_point[0] == math.NaN()) || (intersection_point[1] == math.NaN()) {
		return false
	}

	return e1_bbox.Contains(intersection_point, EPS) &&
		e2_bbox.Contains(intersection_point, EPS)
}

func intersect2DLinesegmentByLine(p0, p1, l0, l1 []float64) []float64 {
	return nil
}

func isFacingUpwards(t Triangle) bool {
	return false
}

func isFacingUpwardsForFace(f Face, vertices []Vertex) bool {
	return false
}

type BBox2d [4]float64

func (b *BBox2d) Add(p interface{}) {

}

func (b *BBox2d) Grow(delta float64) {

}

func (b *BBox2d) Intersects(o BBox2d, epsilon float64) bool {
	return false
}

func (b *BBox2d) Contains(point []float64, epsilon float64) bool {
	return false
}

type BBox3d [6]float64

func (b *BBox3d) to2d() BBox2d {
	return BBox2d{b[0], b[1], b[3], b[4]}
}

func (b *BBox3d) Add(p interface{}) {

}

func (b *BBox3d) Contains(point []float64, epsilon float64) bool {
	return false
}

func (b *BBox3d) Grow(delta float64) {

}

func Clip25dTrianglesTo01Quadrant(vertices []Vertex) {

}
