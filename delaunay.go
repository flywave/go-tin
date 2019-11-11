package tin

import (
	"math"
	"math/rand"
)

type trianglePool []*DelaunayTriangle

type DelaunayTriangle struct {
	Anchor QuadEdge
	Next   *DelaunayTriangle
	pool   trianglePool
}

func NewDelaunayTriangle(p trianglePool) *DelaunayTriangle {
	ptr := &DelaunayTriangle{pool: p}
	p = append(p, ptr)
	return ptr
}

func (t *DelaunayTriangle) init(e QuadEdge) {
	t.reshape(e)
}

func (t *DelaunayTriangle) linkTo(o *DelaunayTriangle) *DelaunayTriangle {
	t.Next = o
	return t
}

func (t *DelaunayTriangle) GetLink() *DelaunayTriangle { return t.Next }

func (t *DelaunayTriangle) GetAnchor() QuadEdge {
	return t.Anchor
}

func (t *DelaunayTriangle) dontAnchor(e QuadEdge) {
	if t.Anchor == e {
		t.Anchor = e.LeftNext()
	}
}

func (t *DelaunayTriangle) reshape(e QuadEdge) {
	t.Anchor = e
	e.SetLeftFace(t)
	e.LeftNext().SetLeftFace(t)
	e.LeftPrev().SetLeftFace(t)
}

func (t *DelaunayTriangle) point1() [2]float64 {
	return t.Anchor.Orig()
}

func (t *DelaunayTriangle) point2() [2]float64 {
	return t.Anchor.Dest()
}

func (t *DelaunayTriangle) point3() [2]float64 {
	return t.Anchor.LeftPrev().Orig()
}

func (t *DelaunayTriangle) clear() {

}

type DelaunayMesh struct {
	QuadEdges        *Pool
	Triangles        trianglePool
	startingQuadEdge QuadEdge
	firstFace        *DelaunayTriangle
}

func (m *DelaunayMesh) makeFace(e QuadEdge) *DelaunayTriangle {
	t := NewDelaunayTriangle(m.Triangles)
	t.init(e)

	m.firstFace = t.linkTo(m.firstFace)
	return t
}

func (m *DelaunayMesh) delete(e QuadEdge) {
	Delete(e)
}

func (m *DelaunayMesh) connect(a QuadEdge, b QuadEdge) QuadEdge {
	return Connect(a, b)
}

func (m *DelaunayMesh) swap(e QuadEdge) {
	f1 := e.LeftFace().(*DelaunayTriangle)
	f2 := e.Sym().LeftFace().(*DelaunayTriangle)

	SwapTriangles(e)

	f1.reshape(e)
	f2.reshape(e.Sym())
}

func rightOf(p [2]float64, e QuadEdge) bool {
	return IsCCW(p, e.Dest(), e.Orig())
}

func leftOf(p [2]float64, e QuadEdge) bool {
	return IsCCW(p, e.Orig(), e.Dest())
}

func (m *DelaunayMesh) ccwBoundary(e QuadEdge) bool {
	return !rightOf(e.OrigPrev().Dest(), e)
}

func sub(x, y [2]float64) []float64 {
	return []float64{(x[0] - y[0]), (x[1] - y[1])}
}

func (m *DelaunayMesh) onQuadEdge(x [2]float64, e QuadEdge) bool {
	t1 := length(sub(x, e.Orig()))
	t2 := length(sub(x, e.Dest()))

	if t1 < EPS || t2 < EPS {
		return true
	}

	t3 := length(sub(e.Orig(), e.Dest()))

	if t1 > t3 || t2 > t3 {
		return false
	}

	l := NewLine(e.Orig(), e.Dest())
	return (math.Abs(l.Eval(x)) < EPS)
}

func (m *DelaunayMesh) initMesh(a, b, c, d [2]float64) {
	ea := New(m.QuadEdges)
	ea.SetEndPoints(a, b)

	eb := New(m.QuadEdges)
	Splice(ea.Sym(), eb)
	eb.SetEndPoints(b, c)

	ec := New(m.QuadEdges)
	Splice(eb.Sym(), ec)
	ec.SetEndPoints(c, d)

	ed := New(m.QuadEdges)
	Splice(ec.Sym(), ed)
	ed.SetEndPoints(d, a)
	Splice(ed.Sym(), ea)

	diag := New(m.QuadEdges)
	Splice(ed.Sym(), diag)
	Splice(eb.Sym(), diag.Sym())
	diag.SetEndPoints(a, c)

	m.startingQuadEdge = ea

	m.firstFace.clear()

	m.makeFace(ea.Sym())
	m.makeFace(ec.Sym())
}

func (m *DelaunayMesh) isInterior(e QuadEdge) bool {
	return (e.LeftNext().LeftNext().LeftNext() == e &&
		e.RightNext().RightNext().RightNext() == e)
}

func isEqual(v, o [2]float64) bool {
	return math.Abs(v[0]-o[0]) < EPS && math.Abs(v[1]-o[1]) < EPS
}

func (m *DelaunayMesh) insert(x [2]float64, tri *DelaunayTriangle) {
	var e QuadEdge
	if tri != nil {
		e = m.locate(x, tri.Anchor)
	} else {
		e = m.locate(x, m.startingQuadEdge)
	}

	if (isEqual(x, e.Orig())) || (isEqual(x, e.Dest())) {
		m.optimize(x, e)
	} else {
		startSpoke := m.spoke(x, e)
		if startSpoke != nil {
			m.optimize(x, startSpoke.Sym())
		}
	}
}

func (m *DelaunayMesh) shouldSwap(x [2]float64, e QuadEdge) bool {
	t := e.OrigPrev()
	return InCircumcircle(e.Orig(), t.Dest(), e.Dest(), x)
}

func triArea(a, b, c [2]float64) float64 {
	return (b[0]-a[0])*(c[1]-a[1]) - (b[1]-a[1])*(c[0]-a[0])
}

func nextRandomNumber() uint32 {
	return rand.Uint32() % math.MaxUint32
}

func (m *DelaunayMesh) locate(x [2]float64, e QuadEdge) QuadEdge {
	t := triArea(x, e.Dest(), e.Orig())

	if t > 0 {
		t = -t
		e = e.Sym()
	}

	for {
		eo := e.OrigNext()
		ed := e.DestPrev()

		to := triArea(x, eo.Dest(), eo.Orig())
		td := triArea(x, ed.Dest(), ed.Orig())

		if td > 0 {
			if to > 0 || (to == 0 && t == 0) {
				m.startingQuadEdge = e
				return e
			}
			t = to
			e = eo
		} else {
			if to > 0 {
				if td == 0 && t == 0 {
					m.startingQuadEdge = e
					return e
				}
				t = td
				e = ed
			} else {
				if t == 0 && !leftOf(eo.Dest(), e) {
					e = e.Sym()
				} else if (nextRandomNumber() & 1) == 0 {
					t = to
					e = eo
				} else {
					t = td
					e = ed
				}
			}
		}
	}
}

func (m *DelaunayMesh) spoke(x [2]float64, e QuadEdge) *QuadEdge {
	var newFaces [4]*DelaunayTriangle
	facedex := 0

	var boundaryQuadEdge *QuadEdge

	lface := e.LeftFace().(*DelaunayTriangle)
	lface.dontAnchor(e)
	newFaces[facedex] = lface
	facedex++

	if m.onQuadEdge(x, e) {
		if m.ccwBoundary(e) {
			boundaryQuadEdge = &e
		} else {
			symLface := e.Sym().LeftFace().(*DelaunayTriangle)
			newFaces[facedex] = symLface
			facedex++

			symLface.dontAnchor(e.Sym())

			e = e.OrigPrev()
			m.delete(e.OrigNext())
		}
	}

	base := New(m.QuadEdges)

	base.SetEndPoints(e.Orig(), x)

	Splice(base, e)

	m.startingQuadEdge = base
	for {
		base = m.connect(e, base.Sym())
		e = base.OrigPrev()
		if e.LeftNext() == m.startingQuadEdge {
			break
		}
	}

	if boundaryQuadEdge != nil {
		m.delete(*boundaryQuadEdge)
	}

	if boundaryQuadEdge != nil {
		base = m.startingQuadEdge.RightPrev()
	} else {
		base = m.startingQuadEdge.Sym()
	}

	for {
		if facedex > 0 {
			facedex--
			newFaces[facedex].reshape(base)
		} else {
			m.makeFace(base)
		}

		base = base.OrigNext()
		if base == m.startingQuadEdge.Sym() {
			break
		}
	}

	return &m.startingQuadEdge
}

func (m *DelaunayMesh) scanTriangle(t *DelaunayTriangle) {
	// noop
}

func (m *DelaunayMesh) optimize(x [2]float64, s QuadEdge) {
	startSpoke := s
	spoke := s

	for {
		e := spoke.LeftNext()
		if m.isInterior(e) && m.shouldSwap(x, e) {
			m.swap(e)
		} else {
			spoke = spoke.OrigNext()
			if spoke == startSpoke {
				break
			}
		}
	}

	spoke = startSpoke

	for {
		e := spoke.LeftNext()
		t := e.LeftFace()

		if t != nil {
			m.scanTriangle(t.(*DelaunayTriangle))
		}

		spoke = spoke.OrigNext()
		if spoke == startSpoke {
			break
		}
	}
}
