package tin

import "math"

const (
	EPS = float64(0.000000001)
)

type Vertex [3]float64
type Normal [3]float64

func (v Vertex) Equal(o Vertex) bool {
	return math.Abs(v[0]-o[0]) < EPS && math.Abs(v[1]-o[1]) < EPS && math.Abs(v[2]-o[2]) < EPS
}

func VertexEqual(v, o Vertex) bool {
	return v.Equal(o)
}

type Triangle [3]Vertex

func (l Triangle) Equal(r Triangle) bool {
	if l[0].Equal(r[0]) && l[1].Equal(r[1]) && l[2].Equal(r[2]) {
		return true
	}
	return false
}

func triangleSemanticEqual(l, r Triangle) bool {
	if l.Equal(r) {
		return true
	}
	var r2 Triangle
	r2[0] = r[1]
	r2[1] = r[2]
	r2[2] = r[0]
	if l.Equal(r2) {
		return true
	}
	var r3 Triangle
	r3[1] = r[0]
	r3[2] = r[1]
	r3[0] = r[2]
	if l.Equal(r3) {
		return true
	}
	return false
}

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
	x1 := p0[0]
	x2 := p1[0]
	x3 := l0[0]
	x4 := l1[0]

	y1 := p0[1]
	y2 := p1[1]
	y3 := l0[1]
	y4 := l1[1]

	denom := (x1-x2)*(y3-y4) - (y1-y2)*(x3-x4)
	if math.Abs(denom) < EPS {
		return []float64{math.NaN(), math.NaN()}
	}

	cX := ((x1*y2-y1*x2)*(x3-x4) -
		(x1-x2)*(x3*y4-y3*x4)) /
		denom
	cY := ((x1*y2-y1*x2)*(y3-y4) -
		(y1-y2)*(x3*y4-y3*x4)) /
		denom

	if cX == -0.0 {
		cX = 0.0
	}
	if cY == -0.0 {
		cY = 0.0
	}
	return []float64{cX, cY}
}

func isFacingUpwardsImpl(t0X, t0Y, t1X, t1Y, t2X, t2Y float64) bool {
	nZ := (t0X-t1X)*(t0Y-t2Y) - (t0X-t2X)*(t0Y-t1Y)
	return nZ >= 0
}

func isFacingUpwards(t Triangle) bool {
	t0X := t[0][0]
	t0Y := t[0][1]

	t1X := t[1][0]
	t1Y := t[1][1]

	t2X := t[2][0]
	t2Y := t[2][1]

	return isFacingUpwardsImpl(t0X, t0Y, t1X, t1Y, t2X, t2Y)
}

func isFacingUpwardsForFace(f Face, vertices []Vertex) bool {
	t0X := vertices[f[0]][0]
	t0Y := vertices[f[0]][1]

	t1X := vertices[f[1]][0]
	t1Y := vertices[f[1]][1]

	t2X := vertices[f[2]][0]
	t2Y := vertices[f[2]][1]

	return isFacingUpwardsImpl(t0X, t0Y, t1X, t1Y, t2X, t2Y)
}

type BBox2d [4]float64

func NewBBox2d() *BBox2d {
	return &BBox2d{-math.MaxFloat64, -math.MaxFloat64, math.MaxFloat64, math.MaxFloat64}
}

func Min(x, y float64) float64 {
	if x < y {
		return x
	}
	return y
}

func Max(x, y float64) float64 {
	if x > y {
		return x
	}
	return y
}

func (b BBox2d) Width() float64 {
	return b[2] - b[0]
}

func (b BBox2d) Height() float64 {
	return b[3] - b[1]
}

func (b *BBox2d) add(p []float64) {
	(*b)[0] = Min((*b)[0], p[0])
	(*b)[1] = Min((*b)[1], p[1])
	(*b)[2] = Max((*b)[2], p[0])
	(*b)[3] = Max((*b)[3], p[1])
}

func (b *BBox2d) Add(p interface{}) {
	switch t := p.(type) {
	case []int:
	case []float32:
		b.add([]float64{float64(t[0]), float64(t[1])})
	case []float64:
		b.add(t)
	case [3]float64:
		b.add(t[:])
	case [2]float64:
		b.add(t[:])
	case Triangle:
		for i := range t {
			b.add(t[i][:])
		}
	}
}

func (b *BBox2d) Grow(delta float64) {
	(*b)[0] -= delta
	(*b)[1] -= delta
	(*b)[2] += delta
	(*b)[3] += delta
}

func epseq(l, r, epsilon float64) bool {
	return math.Abs(l-r) < epsilon
}

func (b BBox2d) IsOnBorder(point []float64, epsilon float64) bool {
	return epseq(point[0], b[0], epsilon) || epseq(point[0], b[2], epsilon) ||
		epseq(point[1], b[1], epsilon) || epseq(point[1], b[3], epsilon)
}

// https://stackoverflow.com/questions/306316/determine-if-two-rectangles-overlap-each-other
func (b BBox2d) Intersects(o BBox2d, epsilon float64) bool {
	if b[1]-epsilon > o[3]+epsilon {
		return false
	}

	if b[3]+epsilon < o[1]-epsilon {
		return false
	}

	if b[2]+epsilon < o[0]-epsilon {
		return false
	}

	if b[0]-epsilon > o[2]+epsilon {
		return false
	}

	return true
}

func (b BBox2d) Contains(point []float64, epsilon float64) bool {
	return (b[0]-epsilon) <= point[0] && (b[1]-epsilon) <= point[1] &&
		(b[2]+epsilon) >= point[0] && (b[3]+epsilon) >= point[1]
}

type BBox3d [6]float64

func NewBBox3d() *BBox3d {
	return &BBox3d{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64}
}

func (b *BBox3d) to2d() BBox2d {
	return BBox2d{b[0], b[1], b[3], b[4]}
}

func (b *BBox3d) add(p []float64) {
	(*b)[0] = Min((*b)[0], p[0])
	(*b)[1] = Min((*b)[1], p[1])
	(*b)[2] = Min((*b)[2], p[2])
	(*b)[3] = Max((*b)[3], p[0])
	(*b)[4] = Max((*b)[4], p[1])
	(*b)[5] = Max((*b)[5], p[2])
}

func (b BBox3d) Width() float64 {
	return b[3] - b[0]
}

func (b BBox3d) Height() float64 {
	return b[4] - b[1]
}

func (b BBox3d) Depth() float64 {
	return b[5] - b[2]
}

func (b *BBox3d) Add(p interface{}) {
	switch t := p.(type) {
	case []int:
	case []float32:
		b.add([]float64{float64(t[0]), float64(t[1]), float64(t[2])})
	case []float64:
		b.add(t)
	case [3]float64:
		b.add(t[:])
	case Triangle:
		for i := range t {
			b.add(t[i][:])
		}
	}
}

func (b BBox3d) Contains(point []float64, epsilon float64) bool {
	return (b[0]-epsilon) <= point[0] && (b[1]-epsilon) <= point[1] &&
		(b[2]-epsilon) <= point[2] &&
		(b[3]+epsilon) >= point[0] && (b[4]+epsilon) >= point[1] &&
		(b[5]+epsilon) >= point[3]
}

func (b *BBox3d) Grow(delta float64) {
	(*b)[0] -= delta
	(*b)[1] -= delta
	(*b)[2] -= delta

	(*b)[3] += delta
	(*b)[4] += delta
	(*b)[5] += delta
}

func absZero(v [3]float64) [3]float64 {
	if v[0] == -0.0 {
		v[0] = 0.0
	}
	if v[1] == -0.0 {
		v[1] = 0.0
	}
	if v[2] == -0.0 {
		v[2] = 0.0
	}
	return v
}

func squared3DDistance(p0, p1 [3]float64) float64 {
	dx := p1[0] - p0[0]
	dy := p1[1] - p0[1]
	dz := p1[2] - p0[2]
	return dx*dx + dy*dy + dz*dz
}

func squared2DDistance(p0, p1 [2]float64) float64 {
	dx := p1[0] - p0[0]
	dy := p1[1] - p0[1]
	return dx*dx + dy*dy
}

/*
	  lp - leftPoints
	  op - otherPoints

	  winding order counter-clockwise = inside
	                  +
	                 /|
	                / |
	               /  |
	              /   |
	             /    |
	            /     |
	           /      |
	          /       |
	         /        |
	        /         |
	l1   s0/          |s1     l0
	 x----*-----------*-------x
	     /            |
	    /             |

lp[0] +              |

	\             |
	  \           |
	    \         |
	      \       |
	        \     |
	          \   |
	            \ |
	              + lp[1]
*/
func Clip25DTriangleByLine(tv []Triangle, triangleIdx int, lorg, ldir [2]float64) []Triangle {
	t := tv[triangleIdx]

	if hasNaNs(t) {
		return tv
	}

	var leftPoints [3][3]float64
	var otherPoints [3][3]float64
	var otherSigns [3]int
	leftPointsCount := 0
	otherPointsCount := 0

	for i := range t {
		d := sign2D(t[i], lorg, ldir)
		if d < 0 {
			leftPoints[leftPointsCount] = t[i]
			leftPointsCount++
		} else {
			otherPoints[otherPointsCount] = t[i]
			otherSigns[otherPointsCount] = d
			otherPointsCount++
		}
	}

	if leftPointsCount == 0 {
		t[0] = [3]float64{math.NaN(), math.NaN(), math.NaN()}
	} else if leftPointsCount == 1 {
		var s0 [3]float64
		var s1 [3]float64

		if otherSigns[0] != 0 {
			s0 = intersect25DLinesegmentByLine(
				leftPoints[0], otherPoints[0], lorg, ldir)
		}
		if otherSigns[1] != 0 {
			s1 = intersect25DLinesegmentByLine(
				leftPoints[0], otherPoints[1], lorg, ldir)
		}

		t[0] = leftPoints[0]
		t[1] = s0
		t[2] = s1
		makeFrontFacing(&t)
	} else if leftPointsCount == 2 {
		if otherSigns[0] == 0 {
			return tv
		}

		s0 := intersect25DLinesegmentByLine(
			otherPoints[0], leftPoints[0], lorg, ldir)
		s1 := intersect25DLinesegmentByLine(
			otherPoints[0], leftPoints[1], lorg, ldir)

		d0d1cmp :=
			compareLength3D(s0, leftPoints[1], s1, leftPoints[0])

		if d0d1cmp >= 0 {
			t[0] = s1
		} else {
			t[0] = s0
		}

		t[1] = leftPoints[0]
		t[2] = leftPoints[1]

		var s2 [3]float64
		if d0d1cmp >= 0 {
			s2 = leftPoints[0]
		} else {
			s2 = leftPoints[1]
		}
		tnew := Triangle{s1, s0, s2}

		makeFrontFacing(&t)
		makeFrontFacing(&tnew)
		tv = append(tv, tnew)
	}
	return tv
}

func Clip25dTrianglesTo01Quadrant(tv []Triangle) []Triangle {
	tvsize := len(tv)
	for i := 0; i < tvsize; i++ {
		Clip25DTriangleByLine(tv, i, [2]float64{0, 0}, [2]float64{1, 0})
	}

	tvsize = len(tv)
	for i := 0; i < tvsize; i++ {
		Clip25DTriangleByLine(tv, i, [2]float64{1, 0}, [2]float64{0, 1})
	}

	tvsize = len(tv)
	for i := 0; i < tvsize; i++ {
		Clip25DTriangleByLine(tv, i, [2]float64{1, 1}, [2]float64{-1, 0})
	}

	tvsize = len(tv)
	for i := 0; i < tvsize; i++ {
		Clip25DTriangleByLine(tv, i, [2]float64{0, 1}, [2]float64{0, -1})
	}

	var new []Triangle
	for i := range tv {
		if !hasNaNs(tv[i]) {
			new = append(new, tv[i])
		}
	}
	return new
}

func length(p []float64) float64 {
	var length float64
	for i := range p {
		length += math.Pow(p[i], 2)
	}
	return math.Sqrt(length)
}

func distance(p0, p1 []float64) float64 {
	if len(p0) == len(p1) {
		var length float64
		for i := range p0 {
			length += (p1[i] - p0[i]) * (p1[i] - p0[i])
		}
		return math.Sqrt(length)
	}
	return math.NaN()
}

// from https://en.wikipedia.org/wiki/Line%E2%80%93line_intersection
func intersect25DLinesegmentByLine(p0, p1 [3]float64, lorg, ldir [2]float64) [3]float64 {
	x1 := p0[0]
	x2 := p1[0]
	x3 := lorg[0]
	x4 := lorg[0] + ldir[0]

	y1 := p0[1]
	y2 := p1[1]
	y3 := lorg[1]
	y4 := lorg[1] + ldir[1]

	denom := (x1-x2)*(y3-y4) - (y1-y2)*(x3-x4)
	if math.Abs(denom) < EPS {
		return [3]float64{math.NaN(), math.NaN(), math.NaN()}
	}

	var c [3]float64
	c[0] = ((x1*y2-y1*x2)*(x3-x4) -
		(x1-x2)*(x3*y4-y3*x4)) /
		denom
	c[1] = ((x1*y2-y1*x2)*(y3-y4) -
		(y1-y2)*(x3*y4-y3*x4)) /
		denom

	dp0p1 := distance(p0[0:2], p1[0:2])
	m := (p1[2] - p0[2]) / dp0p1
	n := p0[2]

	dp0c := distance(p0[0:2], c[0:2])
	if dp0c < 0.0-EPS || dp0c > (dp0p1+EPS) {
		return [3]float64{math.NaN(), math.NaN(), math.NaN()}
	}

	c[2] = m*dp0c + n
	return absZero(c)
}

func sign2D(p [3]float64, lorg, ldir [2]float64) int {
	if ldir[0] == 0.0 {
		var directionSign int
		if ldir[1] > 0.0 {
			directionSign = -1
		} else {
			directionSign = 1
		}
		if p[0] < lorg[0] {
			return directionSign
		} else if p[0] > lorg[0] {
			return -directionSign
		} else {
			return 0
		}
	} else if ldir[1] == 0.0 {
		var directionSign int
		if ldir[0] > 0.0 {
			directionSign = -1
		} else {
			directionSign = 1
		}
		if p[1] < lorg[1] {
			return -directionSign
		} else if p[1] > lorg[1] {
			return directionSign
		} else {
			return 0
		}
	} else {
		d := (p[0]-lorg[0])*(ldir[1]) - (p[1]-lorg[1])*(ldir[0])
		if d < EPS {
			return -1
		} else if d > EPS {
			return 1
		} else {
			return 0
		}
	}
}

func compareLength3D(a1, a2, b1, b2 [3]float64) int {
	daSq := squared3DDistance(a1, a2)
	dbSq := squared3DDistance(b1, b2)

	if daSq < dbSq {
		return -1
	} else if daSq == dbSq {
		return 0
	} else {
		return 1
	}
}

func compareLength2D(a1, a2, b1, b2 [2]float64) int {
	daSq := squared2DDistance(a1, a2)
	dbSq := squared2DDistance(b1, b2)

	if daSq < dbSq {
		return -1
	} else if daSq == dbSq {
		return 0
	} else {
		return 1
	}
}

func hasNaNs(t Triangle) bool {
	for i := 0; i < 3; i++ {
		if t[i][0] == math.NaN() || t[i][1] == math.NaN() || t[i][2] == math.NaN() {
			return true
		}
	}
	return false
}

func isFrontFacing(t Triangle) bool {
	t0X := t[0][0]
	t0Y := t[0][1]
	nZ :=
		(t0X-t[1][0])*(t0Y-t[2][1]) - (t0X-t[2][0])*(t0Y-t[1][1])
	return nZ >= 0
}

func makeFrontFacing(t *Triangle) {
	if !isFrontFacing(*t) {
		tem := t[0]
		t[0] = t[1]
		t[1] = tem
	}
}

type Line [3]float64

func NewLine(a, b [2]float64) *Line {
	t := [2]float64{a[0] - b[0], a[1] - b[1]}
	l := length(t[:])

	x := t[1] / l
	y := -t[0] / l
	return &Line{x, y, -(x*a[0] + y*a[1])}
}

func (l Line) Eval(p [2]float64) float64 {
	return (l[0]*p[0] + l[1]*p[1] + l[2])
}

type Plane [3]float64

func NewPlane(p, q, r [3]float64) *Plane {
	ux := q[0] - p[0]
	uy := q[1] - p[1]
	uz := q[2] - p[2]

	vx := r[0] - p[0]
	vy := r[1] - p[1]
	vz := r[2] - p[2]

	den := ux*vy - uy*vx

	_a := (uz*vy - uy*vz) / den
	_b := (ux*vz - uz*vx) / den

	a := _a
	b := _b
	c := p[2] - _a*p[0] - _b*p[1]
	return &Plane{a, b, c}
}

func (p Plane) Eval(x, y float64) float64 {
	return p[0]*x + p[1]*y + p[2]
}
