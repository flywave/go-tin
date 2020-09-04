package tin

func Splice(a, b *QuadEdge) {
	alpha := a.OrigNext().Rot()
	beta := b.OrigNext().Rot()

	t1 := b.OrigNext()
	t2 := a.OrigNext()
	t3 := beta.OrigNext()
	t4 := alpha.OrigNext()

	a.next = t1
	b.next = t2
	alpha.next = t3
	beta.next = t4
}

func Connect(a, b *QuadEdge) *QuadEdge {
	e := New(a.pool)
	e.Init()
	e.SetOrig(a.Dest())
	e.SetDest(b.Orig())
	Splice(e, a.LeftNext())
	Splice(e.Sym(), b)
	return e
}

func SwapTriangles(e *QuadEdge) {
	a := e.OrigPrev()
	b := e.Sym().OrigPrev()
	Splice(e, a)
	Splice(e.Sym(), b)
	Splice(e, a.LeftNext())
	Splice(e.Sym(), b.LeftNext())
	e.SetOrig(a.Dest())
	e.SetDest(b.Dest())
}
