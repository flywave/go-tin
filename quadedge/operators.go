package quadedge

func Splice(a, b Edge) {
	p := a.pool
	alpha := p.next[a.id].rot()
	beta := p.next[b.id].rot()

	p.next[a.id], p.next[b.id] = p.next[b.id], p.next[a.id]

	p.next[alpha], p.next[beta] = p.next[beta], p.next[alpha]
}

func Connect(a, b Edge) Edge {
	e := New(a.pool)
	e.SetOrig(a.Dest())
	e.SetDest(b.Orig())
	Splice(e, a.LeftNext())
	Splice(e.Sym(), b)
	return e
}

func SwapTriangles(e Edge) {
	a := e.OrigPrev()
	b := e.Sym().OrigPrev()
	Splice(e, a)
	Splice(e.Sym(), b)
	Splice(e, a.LeftNext())
	Splice(e.Sym(), b.LeftNext())
	e.SetOrig(a.Dest())
	e.SetDest(b.Dest())
}
