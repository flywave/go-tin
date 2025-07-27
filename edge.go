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
	if e == nil || e.qnext == nil {
		return nil
	}
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
	rot := e.Rot()
	if rot == nil {
		return nil
	}
	return rot.OrigNext().Tor()
}

func (e *QuadEdge) OrigPrev() *QuadEdge {
	rot := e.Rot()
	if rot == nil {
		return nil
	}
	return rot.Next().Rot()
}

func (e *QuadEdge) DestNext() *QuadEdge {
	sym := e.Sym()
	if sym == nil {
		return nil
	}
	return sym.Next().Sym()
}

func (e *QuadEdge) DestPrev() *QuadEdge {
	tor := e.Tor()
	if tor == nil {
		return nil
	}
	return tor.Next().Tor()
}

func (e *QuadEdge) LeftNext() *QuadEdge {
	tor := e.Tor()
	if tor == nil {
		return nil
	}
	return tor.Next().Rot()
}

func (e *QuadEdge) LeftPrev() *QuadEdge {
	return e.Next().Sym()
}

func (e *QuadEdge) RightPrev() *QuadEdge {
	return e.Sym().Next()
}

func (e *QuadEdge) Orig() [2]float64 {
	return e.data
}

func (e *QuadEdge) Dest() [2]float64 {
	sym := e.Sym()
	if sym == nil {
		return [2]float64{}
	}
	return sym.data
}

func (e *QuadEdge) SetOrig(d [2]float64) {
	e.data = d
}

func (e *QuadEdge) SetDest(d [2]float64) {
	sym := e.Sym()
	if sym != nil {
		sym.data = d
	}
}

func (e *QuadEdge) LeftFace() *DelaunayTriangle {
	return e.lface
}

func (e *QuadEdge) SetLeftFace(f *DelaunayTriangle) {
	e.lface = f
}

func (e *QuadEdge) clear() {
	if e != nil && e.pool != nil && e.index > 0 {
		e.pool.clear(e.index)
		e.pool = nil
		e.index = -1
	}
}

func (e *QuadEdge) SetEndPoints(org [2]float64, dest [2]float64) {
	e.data = org
	sym := e.Sym()
	if sym != nil {
		sym.data = dest
	}
}

func (e *QuadEdge) recycle() {
	e.clear()
}

// 优化后的回收方法
func (e *QuadEdge) RecycleNext() {
	if e == nil || e.index <= 0 {
		return
	}

	// 收集四元组的所有边
	edges := []*QuadEdge{e, e.qnext, e.qnext.qnext, e.qprev}

	// 清除所有边的连接关系
	for _, edge := range edges {
		if edge != nil {
			edge.qnext = nil
			edge.qprev = nil
			edge.next = nil
			edge.lface = nil
		}
	}

	// 批量回收边
	for _, edge := range edges {
		if edge != nil {
			edge.recycle()
		}
	}
}
