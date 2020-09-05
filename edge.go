package tin

type QuadEdge struct {
	pool  *Pool
	qnext *QuadEdge
	qprev *QuadEdge
	next  *QuadEdge
	data  [2]float64
	lface *DelaunayTriangle
	index int
}

type edgeID uint32

const (
	Nil              = 0xFFFFFFFF
	canonical edgeID = 0xFFFFFFFC
	quad      edgeID = 0x00000003
)

func (e *QuadEdge) Init() {
	e1 := New(e.pool)
	e2 := New(e.pool)
	e3 := New(e.pool)

	e1.Init2(e)
	e2.Init2(e1)
	e3.Init2(e2)

	e.qprev = e3
	e3.qnext = e

	e.next = e
	e1.next = e3
	e2.next = e2
	e3.next = e1
}

func (e *QuadEdge) Init2(ed *QuadEdge) {
	e.qprev = ed
	ed.qnext = e
}

func (e *QuadEdge) Pool() *Pool {
	return e.pool
}

func (e *QuadEdge) Rot() *QuadEdge {
	return e.qnext
}

func (e *QuadEdge) Sym() *QuadEdge {
	return e.qnext.qnext
}

func (e *QuadEdge) Tor() *QuadEdge {
	return e.qprev
}

func (e *QuadEdge) OrigNext() *QuadEdge {
	return e.next
}

func (e *QuadEdge) Next() *QuadEdge {
	return e.next
}

func (e *QuadEdge) RightNext() *QuadEdge {
	return e.Rot().OrigNext().Tor()
}

func (e *QuadEdge) OrigPrev() *QuadEdge { return e.Rot().Next().Rot() }
func (e *QuadEdge) DestNext() *QuadEdge { return e.Sym().Next().Sym() }
func (e *QuadEdge) DestPrev() *QuadEdge { return e.Tor().Next().Tor() }

func (e *QuadEdge) LeftNext() *QuadEdge { return e.Tor().Next().Rot() }

func (e *QuadEdge) LeftPrev() *QuadEdge { return e.Next().Sym() }

func (e *QuadEdge) RightPrev() *QuadEdge { return e.Sym().Next() }

func (e *QuadEdge) Orig() [2]float64 { return e.data }

func (e *QuadEdge) Dest() [2]float64 { return e.Sym().data }

func (e *QuadEdge) SetOrig(d [2]float64) { e.data = d }

func (e *QuadEdge) SetDest(d [2]float64) { e.Sym().data = d }

func (e *QuadEdge) LeftFace() *DelaunayTriangle { return e.lface }

func (e *QuadEdge) SetLeftFace(f *DelaunayTriangle) { e.lface = f }

func (e *QuadEdge) clear() {
	if e != nil && e.pool != nil && e.index > 0 {
		e.pool.clear(e.index)
		e.pool = nil
		e.index = -1
	}
}

func (e *QuadEdge) SetEndPoints(org [2]float64, dest [2]float64) {
	e.data = org
	e.Sym().data = dest
}

func (e *QuadEdge) recycle() {
	if e.pool != nil {
		e.clear()
	}
}

func (e *QuadEdge) RecycleNext() {
	if e.index > 0 {
		e1 := e.qnext
		e2 := e.qnext.qnext
		e3 := e.qprev

		e1.qnext.clear()
		e2.qnext.clear()
		e3.qnext.clear()

		e1.RecycleNext()
		e1.recycle()
		e2.RecycleNext()
		e2.recycle()
		e3.RecycleNext()
		e3.recycle()
	}
}
