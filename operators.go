package tin

// Splice - 优化后的边连接操作
func Splice(a, b *QuadEdge) {
	// 添加安全保护
	if a == nil || b == nil {
		return
	}

	// 获取相关边
	alpha := a.OrigNext().Rot()
	beta := b.OrigNext().Rot()

	// 添加安全保护
	if alpha == nil || beta == nil {
		return
	}

	// 临时存储
	t1 := b.OrigNext()
	t2 := a.OrigNext()
	t3 := beta.OrigNext()
	t4 := alpha.OrigNext()

	// 更新连接
	a.next = t1
	b.next = t2
	alpha.next = t3
	beta.next = t4
}

// Connect - 优化后的边连接创建
func Connect(a, b *QuadEdge) *QuadEdge {
	if a == nil || b == nil {
		return nil
	}

	e := New(a.pool)
	if e == nil {
		return nil
	}

	e.Init()
	e.SetOrig(a.Dest())
	e.SetDest(b.Orig())

	// 使用安全版本的Splice
	Splice(e, a.LeftNext())
	Splice(e.Sym(), b)

	return e
}

// SwapTriangles - 优化后的三角形交换
func SwapTriangles(e *QuadEdge) {
	if e == nil || e.Sym() == nil {
		return
	}

	a := e.OrigPrev()
	b := e.Sym().OrigPrev()

	// 添加安全保护
	if a == nil || b == nil {
		return
	}

	Splice(e, a)
	Splice(e.Sym(), b)
	Splice(e, a.LeftNext())
	Splice(e.Sym(), b.LeftNext())

	e.SetOrig(a.Dest())
	e.SetDest(b.Dest())
}
