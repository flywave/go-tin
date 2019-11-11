package tin

import "math"

type Data struct {
	data  [2]float64
	lface interface{}
}

var (
	NAN_DATA = Data{data: [2]float64{math.NaN(), math.NaN()}, lface: nil}
)

type Pool struct {
	next     []edgeID
	data     []Data
	marks    []uint32
	free     []edgeID
	capacity uint32
	nextMark uint32
}

func NewPool(capacity uint32) *Pool {
	return &Pool{
		next:     make([]edgeID, 0, 4*capacity),
		data:     make([]Data, 0, 4*capacity),
		marks:    make([]uint32, 0, capacity),
		nextMark: 1,
	}
}

func New(p *Pool) QuadEdge {
	var e edgeID

	if len(p.free) > 0 {
		e = p.free[0]
		p.free[0] = p.free[len(p.free)-1]
		p.free = p.free[:len(p.free)-1]
	} else {
		e = edgeID(len(p.next))
		p.next = append(p.next, []edgeID{Nil, Nil, Nil, Nil}...)
		p.data = append(p.data, []Data{NAN_DATA, NAN_DATA, NAN_DATA, NAN_DATA}...)
		p.marks = append(p.marks, 0)
	}

	p.next[e] = e
	p.data[e] = NAN_DATA
	p.next[e.sym()] = e.sym()
	p.data[e.sym()] = NAN_DATA
	p.next[e.rot()] = e.tor()
	p.data[e.rot()] = NAN_DATA
	p.next[e.tor()] = e.rot()
	p.data[e.tor()] = NAN_DATA

	return QuadEdge{pool: p, id: e}
}

func Delete(e QuadEdge) {
	p := e.pool
	f := e.Sym()
	if p.next[e.id] != e.id {
		Splice(e, e.OrigPrev())
	}
	if p.next[f.id] != f.id {
		Splice(f, f.OrigPrev())
	}

	p.next[e.id] = Nil
	p.data[e.id] = NAN_DATA
	p.next[e.id.rot()] = Nil
	p.data[e.id.rot()] = NAN_DATA
	p.next[e.id.sym()] = Nil
	p.data[e.id.sym()] = NAN_DATA
	p.next[e.id.tor()] = Nil
	p.data[e.id.tor()] = NAN_DATA

	p.marks[e.id>>2] = 0

	p.free = append(p.free, e.id)
}
