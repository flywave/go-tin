package tin

import (
	"github.com/flywave/go-tin/quadedge"
)

type trianglePool []*DelaunayTriangle

const (
	Nil                  = 0xFFFFFFFF
	canonical triangleID = 0xFFFFFFFC
	quad      triangleID = 0x00000003
)

type triangleID uint32

type DelaunayTriangle struct {
	Anchor quadedge.Edge
	Next   *DelaunayTriangle
	id     triangleID
	pool   *trianglePool
}

func (e triangleID) rot() triangleID {
	return (e & canonical) + ((e + 1) & quad)
}

func (e triangleID) sym() triangleID {
	return (e & canonical) + ((e + 2) & quad)
}

func (e triangleID) tor() triangleID {
	return (e & canonical) + ((e + 3) & quad)
}

func NewDelaunayTriangle(p *trianglePool) *DelaunayTriangle {
	return nil
}

func DeleteDelaunayTriangle(e *trianglePool) {

}

func (t *DelaunayTriangle) init(e quadedge.Edge) {
	t.reshape(e)
}

func (t *DelaunayTriangle) linkTo(o *DelaunayTriangle) *DelaunayTriangle {
	t.Next = o
	return t
}

func (t *DelaunayTriangle) GetLink() *DelaunayTriangle { return t.Next }

func (t *DelaunayTriangle) GetAnchor() quadedge.Edge {
	return t.Anchor
}

func (t *DelaunayTriangle) dontAnchor(e quadedge.Edge) {
	if t.Anchor == e {
		t.Anchor = e.LeftNext()
	}
}

func (t *DelaunayTriangle) reshape(e quadedge.Edge) {
	t.Anchor = e
	e.SetLeft(uint32(t.id))
	e.LeftNext().SetLeft(uint32(t.id))
	e.LeftPrev().SetLeft(uint32(t.id))
}

func (t *DelaunayTriangle) point1() uint32 { return t.Anchor.Orig() }
func (t *DelaunayTriangle) point2() uint32 { return t.Anchor.Dest() }
func (t *DelaunayTriangle) point3() uint32 { return t.Anchor.LeftPrev().Orig() }

type DelaunayMesh struct {
	Edges     *quadedge.Pool
	Triangles *trianglePool
}

func (m *DelaunayMesh) initMesh(a, b, c, d [2]float64) {

}

func (m *DelaunayMesh) insert(x [2]float64, tri interface{}) {

}
