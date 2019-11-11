package tin

import (
	"fmt"
	"strconv"
)

type QuadEdge struct {
	pool *Pool
	id   edgeID
}

type edgeID uint32

const (
	Nil              = 0xFFFFFFFF
	canonical edgeID = 0xFFFFFFFC
	quad      edgeID = 0x00000003
)

func (e QuadEdge) String() string {
	if e.id == Nil {
		return "[no edge]"
	}
	on := e.pool.next[e.id]
	o := "orig:" +
		datastring(e.pool, e.id) +
		"->" + on.String()
	rn := e.pool.next[e.id.rot()]
	r := " right:" +
		datastring(e.pool, e.id.rot()) +
		"->" + rn.String()
	dn := e.pool.next[e.id.sym()]
	d := " dest:" +
		datastring(e.pool, e.id.sym()) +
		"->" + dn.String()
	ln := e.pool.next[e.id.tor()]
	l := " left:" +
		datastring(e.pool, e.id.tor()) +
		"->" + ln.String()
	return e.id.String() + "=[" + o + r + d + l + "]"
}

func (e edgeID) String() string {
	if e == Nil {
		return "no_edge"
	}
	var s string
	switch e & quad {
	case 0:
		s = "o"
	case 1:
		s = "r"
	case 2:
		s = "d"
	case 3:
		s = "l"
	default:
		s = "error"
	}
	return strconv.Itoa(int(e>>2)) + "" + s
}

func datastring(p *Pool, e edgeID) string {
	d := p.data[e]
	if d == NAN_DATA {
		return ""
	}
	return fmt.Sprintf("%.5f-%.5f", d.data[0], d.data[1])
}

func (e QuadEdge) Pool() *Pool {
	return e.pool
}

func (e edgeID) rot() edgeID {
	return (e & canonical) + ((e + 1) & quad)
}

func (e edgeID) sym() edgeID {
	return (e & canonical) + ((e + 2) & quad)
}

func (e edgeID) tor() edgeID {
	return (e & canonical) + ((e + 3) & quad)
}

func (e QuadEdge) Canonical() QuadEdge {
	e.id = e.id & canonical
	return e
}

func (e QuadEdge) Rot() QuadEdge {
	e.id = (e.id & canonical) + ((e.id + 1) & quad)
	return e
}

func (e QuadEdge) Sym() QuadEdge {
	e.id = (e.id & canonical) + ((e.id + 2) & quad)
	return e
}

func (e QuadEdge) Tor() QuadEdge {
	e.id = (e.id & canonical) + ((e.id + 3) & quad)
	return e
}

func (e QuadEdge) OrigNext() QuadEdge {
	e.id = e.pool.next[e.id]
	return e
}

func (e QuadEdge) RightNext() QuadEdge {
	e.id = e.pool.next[e.id.rot()].tor()
	return e
}

func (e QuadEdge) DestNext() QuadEdge {
	e.id = e.pool.next[e.id.sym()].sym()
	return e
}

func (e QuadEdge) LeftNext() QuadEdge {
	e.id = e.pool.next[e.id.tor()].rot()
	return e
}

func (e QuadEdge) OrigPrev() QuadEdge {
	e.id = e.pool.next[e.id.rot()].rot()
	return e
}

func (e QuadEdge) RightPrev() QuadEdge {
	e.id = e.pool.next[e.id.sym()]
	return e
}

func (e QuadEdge) DestPrev() QuadEdge {
	e.id = e.pool.next[e.id.tor()].tor()
	return e
}

func (e QuadEdge) LeftPrev() QuadEdge {
	e.id = e.pool.next[e.id].sym()
	return e
}

func (e QuadEdge) OrigLoop(visit func(e QuadEdge)) {
	f := e.id
	for e.id != Nil {
		visit(e)
		e.id = e.pool.next[e.id]
		if e.id == f {
			return
		}
	}
}

func (e QuadEdge) RightLoop(visit func(e QuadEdge)) {
	f := e.id
	for e.id != Nil {
		visit(e)
		e.id = e.pool.next[e.id.rot()].tor()
		if e.id == f {
			return
		}
	}
}

func (e QuadEdge) DestLoop(visit func(e QuadEdge)) {
	f := e.id
	for e.id != Nil {
		visit(e)
		e.id = e.pool.next[e.id.sym()].sym()
		if e.id == f {
			return
		}
	}
}

func (e QuadEdge) LeftLoop(visit func(e QuadEdge)) {
	f := e.id
	for e.id != Nil {
		visit(e)
		e.id = e.pool.next[e.id.tor()].rot()
		if e.id == f {
			return
		}
	}
}

func (e QuadEdge) SameRing(o QuadEdge) bool {
	f := e.id
	for e.id != Nil {
		if e.id == o.id {
			return true
		}
		e.id = e.pool.next[e.id]
		if e.id == f {
			return false
		}
	}
	return false
}

func (e QuadEdge) Orig() [2]float64 {
	return e.pool.data[e.id].data
}

func (e QuadEdge) SetOrig(data [2]float64) {
	e.pool.data[e.id].data = data
}

func (e QuadEdge) Right() [2]float64 {
	return e.pool.data[e.id.rot()].data
}

func (e QuadEdge) SetRight(data [2]float64) {
	e.pool.data[e.id.rot()].data = data
}

func (e QuadEdge) Dest() [2]float64 {
	return e.pool.data[e.id.sym()].data
}

func (e QuadEdge) SetDest(data [2]float64) {
	e.pool.data[e.id.sym()].data = data
}

func (e QuadEdge) Left() [2]float64 {
	return e.pool.data[e.id.tor()].data
}

func (e QuadEdge) SetLeft(data [2]float64) {
	e.pool.data[e.id.tor()].data = data
}

func (e QuadEdge) mark() uint32 {
	return e.pool.marks[e.id>>2]
}

func (e QuadEdge) setMark(mark uint32) {
	e.pool.marks[e.id>>2] = mark
}

func (e QuadEdge) Walk(visit func(e QuadEdge)) {
	if e.id == Nil {
		return
	}
	m := e.pool.nextMark
	e.pool.nextMark++
	if e.pool.nextMark == 0 {
		e.pool.nextMark = 1
	}
	e.pool.walk(e.id, visit, m)
}

func (e QuadEdge) SetEndPoints(org, dest [2]float64) {
	e.pool.data[e.id.tor()].data = org
	e.pool.data[e.id.sym()].data = dest
}

func (e QuadEdge) SetLeftFace(t interface{}) {
	e.pool.data[e.id].lface = t
}

func (e QuadEdge) LeftFace() interface{} {
	return e.pool.data[e.id].lface
}

func (p *Pool) walk(eid edgeID, visit func(e QuadEdge), m uint32) {
	for p.marks[eid>>2] != m {
		visit(QuadEdge{pool: p, id: eid})
		p.marks[eid>>2] = m
		p.walk(p.next[eid.sym()], visit, m)
		eid = p.next[eid]
	}
}
