package quadedge

import (
	"strconv"
)

type Edge struct {
	pool *Pool
	id   edgeID
}

type edgeID uint32

const (
	Nil              = 0xFFFFFFFF
	canonical edgeID = 0xFFFFFFFC
	quad      edgeID = 0x00000003
)

func (e Edge) String() string {
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
	if d == Nil {
		return ""
	}
	return strconv.Itoa(int(d))
}

func (e Edge) Pool() *Pool {
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

func (e Edge) Canonical() Edge {
	e.id = e.id & canonical
	return e
}

func (e Edge) Rot() Edge {
	e.id = (e.id & canonical) + ((e.id + 1) & quad)
	return e
}

func (e Edge) Sym() Edge {
	e.id = (e.id & canonical) + ((e.id + 2) & quad)
	return e
}

func (e Edge) Tor() Edge {
	e.id = (e.id & canonical) + ((e.id + 3) & quad)
	return e
}

func (e Edge) OrigNext() Edge {
	e.id = e.pool.next[e.id]
	return e
}

func (e Edge) RightNext() Edge {
	e.id = e.pool.next[e.id.rot()].tor()
	return e
}

func (e Edge) DestNext() Edge {
	e.id = e.pool.next[e.id.sym()].sym()
	return e
}

func (e Edge) LeftNext() Edge {
	e.id = e.pool.next[e.id.tor()].rot()
	return e
}

func (e Edge) OrigPrev() Edge {
	e.id = e.pool.next[e.id.rot()].rot()
	return e
}

func (e Edge) RightPrev() Edge {
	e.id = e.pool.next[e.id.sym()]
	return e
}

func (e Edge) DestPrev() Edge {
	e.id = e.pool.next[e.id.tor()].tor()
	return e
}

func (e Edge) LeftPrev() Edge {
	e.id = e.pool.next[e.id].sym()
	return e
}

func (e Edge) OrigLoop(visit func(e Edge)) {
	f := e.id
	for e.id != Nil {
		visit(e)
		e.id = e.pool.next[e.id]
		if e.id == f {
			return
		}
	}
}

func (e Edge) RightLoop(visit func(e Edge)) {
	f := e.id
	for e.id != Nil {
		visit(e)
		e.id = e.pool.next[e.id.rot()].tor()
		if e.id == f {
			return
		}
	}
}

func (e Edge) DestLoop(visit func(e Edge)) {
	f := e.id
	for e.id != Nil {
		visit(e)
		e.id = e.pool.next[e.id.sym()].sym()
		if e.id == f {
			return
		}
	}
}

func (e Edge) LeftLoop(visit func(e Edge)) {
	f := e.id
	for e.id != Nil {
		visit(e)
		e.id = e.pool.next[e.id.tor()].rot()
		if e.id == f {
			return
		}
	}
}

func (e Edge) SameRing(o Edge) bool {
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

func (e Edge) Orig() uint32 {
	return e.pool.data[e.id]
}

func (e Edge) SetOrig(data uint32) {
	e.pool.data[e.id] = data
}

func (e Edge) Right() uint32 {
	return e.pool.data[e.id.rot()]
}

func (e Edge) SetRight(data uint32) {
	e.pool.data[e.id.rot()] = data
}

func (e Edge) Dest() uint32 {
	return e.pool.data[e.id.sym()]
}

func (e Edge) SetDest(data uint32) {
	e.pool.data[e.id.sym()] = data
}

func (e Edge) Left() uint32 {
	return e.pool.data[e.id.tor()]
}

func (e Edge) SetLeft(data uint32) {
	e.pool.data[e.id.tor()] = data
}

func (e Edge) mark() uint32 {
	return e.pool.marks[e.id>>2]
}

func (e Edge) setMark(mark uint32) {
	e.pool.marks[e.id>>2] = mark
}

func (e Edge) Walk(visit func(e Edge)) {
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

func (p *Pool) walk(eid edgeID, visit func(e Edge), m uint32) {
	for p.marks[eid>>2] != m {
		visit(Edge{pool: p, id: eid})
		p.marks[eid>>2] = m
		p.walk(p.next[eid.sym()], visit, m)
		eid = p.next[eid]
	}
}
