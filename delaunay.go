package tin

import (
	"math"
	"math/rand"
)

type DelaunayTriangle struct {
	Anchor *QuadEdge
	Next   *DelaunayTriangle
	pool   *Pool
	index  int
}

func NewDelaunayTriangle(p *Pool) *DelaunayTriangle {
	ptr := &DelaunayTriangle{pool: p, index: len(p.Values)}
	p.Values = append(p.Values, ptr)
	return ptr
}

func (t *DelaunayTriangle) init(e *QuadEdge) {
	t.reshape(e)
}

func (t *DelaunayTriangle) linkTo(o *DelaunayTriangle) *DelaunayTriangle {
	t.Next = o
	return t
}

func (t *DelaunayTriangle) GetLink() *DelaunayTriangle { return t.Next }

func (t *DelaunayTriangle) GetAnchor() *QuadEdge {
	return t.Anchor
}

func (t *DelaunayTriangle) dontAnchor(e *QuadEdge) {
	// 添加 nil 检查防止空指针异常
	if t == nil || e == nil {
		return
	}
	if t.Anchor == e {
		t.Anchor = e.LeftNext()
	}
}

func (t *DelaunayTriangle) reshape(e *QuadEdge) {
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

type DelaunayMesh struct {
	QuadEdges        *Pool
	Triangles        *Pool
	startingQuadEdge *QuadEdge
	firstFace        *DelaunayTriangle
	scanTriangle     func(*DelaunayTriangle)
}

func (m *DelaunayMesh) makeFace(e *QuadEdge) *DelaunayTriangle {
	t := NewDelaunayTriangle(m.Triangles)
	t.init(e)

	m.firstFace = t.linkTo(m.firstFace)
	return t
}

func (m *DelaunayMesh) delete(e *QuadEdge) {
	Splice(e, e.OrigPrev())
	Splice(e.Sym(), e.Sym().OrigPrev())
	e.RecycleNext()
	e.recycle()
}

func (m *DelaunayMesh) connect(a *QuadEdge, b *QuadEdge) *QuadEdge {
	return Connect(a, b)
}

func (m *DelaunayMesh) swap(e *QuadEdge) {
	if e == nil || e.Sym() == nil {
		return
	}
	f1 := e.LeftFace()
	f2 := e.Sym().LeftFace()

	SwapTriangles(e)

	// Add nil checks before reshaping
	if f1 != nil {
		f1.reshape(e)
	}
	if f2 != nil {
		f2.reshape(e.Sym())
	}
}

func rightOf(p [2]float64, e *QuadEdge) bool {
	return IsCCW(p, e.Dest(), e.Orig())
}

func leftOf(p [2]float64, e *QuadEdge) bool {
	return IsCCW(p, e.Orig(), e.Dest())
}

func (m *DelaunayMesh) ccwBoundary(e *QuadEdge) bool {
	return !rightOf(e.OrigPrev().Dest(), e)
}

func sub(x, y [2]float64) [2]float64 {
	return [2]float64{x[0] - y[0], x[1] - y[1]}
}

func sqLength(v [2]float64) float64 {
	return v[0]*v[0] + v[1]*v[1]
}

func (m *DelaunayMesh) onQuadEdge(x [2]float64, e *QuadEdge) bool {
	// 使用平方距离比较避免开方运算
	orig := e.Orig()
	dest := e.Dest()

	// 提前检查点重合
	if sqLength(sub(x, orig)) < EPS*EPS || sqLength(sub(x, dest)) < EPS*EPS {
		return true
	}

	// 检查是否在延长线上
	edgeVec := sub(dest, orig)
	edgeLenSq := sqLength(edgeVec)
	toOrig := sub(x, orig)
	toDest := sub(x, dest)

	// 如果点在边的延长线上，但不在线段范围内
	if sqLength(toOrig) > edgeLenSq || sqLength(toDest) > edgeLenSq {
		return false
	}

	// 精确的共线检查
	l := NewLine(orig, dest)
	return math.Abs(l.Eval(x)) < EPS
}

func (m *DelaunayMesh) InitMeshFromBBox(bb BBox2d) {
	a := [2]float64{bb[0], bb[1]}
	d := [2]float64{bb[2], bb[3]}
	b := [2]float64{bb[2], bb[1]}
	c := [2]float64{bb[0], bb[3]}
	m.initMesh(a, b, c, d)
}

func (m *DelaunayMesh) initMesh(a, b, c, d [2]float64) {
	ea := New(m.QuadEdges)
	ea.Init()
	ea.SetEndPoints(a, b)

	eb := New(m.QuadEdges)
	eb.Init()
	Splice(ea.Sym(), eb)
	eb.SetEndPoints(b, c)

	ec := New(m.QuadEdges)
	ec.Init()
	Splice(eb.Sym(), ec)
	ec.SetEndPoints(c, d)

	ed := New(m.QuadEdges)
	ed.Init()
	Splice(ec.Sym(), ed)
	ed.SetEndPoints(d, a)
	Splice(ed.Sym(), ea)

	diag := New(m.QuadEdges)
	diag.Init()
	Splice(ed.Sym(), diag)
	Splice(eb.Sym(), diag.Sym())
	diag.SetEndPoints(a, c)

	m.startingQuadEdge = ea

	m.firstFace = nil

	m.makeFace(ea.Sym())
	m.makeFace(ec.Sym())
}

func (m *DelaunayMesh) isInterior(e *QuadEdge) bool {
	return (e.LeftNext().LeftNext().LeftNext() == e &&
		e.RightNext().RightNext().RightNext() == e)
}

func isEqual(v, o [2]float64) bool {
	return math.Abs(v[0]-o[0]) < EPS && math.Abs(v[1]-o[1]) < EPS
}

func (m *DelaunayMesh) insert(x [2]float64, tri *DelaunayTriangle) {
	var e *QuadEdge
	if tri != nil {
		e = m.locate(x, tri.Anchor)
	} else {
		e = m.locate(x, m.startingQuadEdge)
	}

	if isEqual(x, e.Orig()) || isEqual(x, e.Dest()) {
		m.optimize(x, e)
	} else {
		startSpoke := m.spoke(x, e)
		if startSpoke != nil {
			m.optimize(x, startSpoke.Sym())
		}
	}
}

func (m *DelaunayMesh) shouldSwap(x [2]float64, e *QuadEdge) bool {
	t := e.OrigPrev()
	return InCircumcircle(e.Orig(), t.Dest(), e.Dest(), x)
}

func triArea(a, b, c [2]float64) float64 {
	return (b[0]-a[0])*(c[1]-a[1]) - (b[1]-a[1])*(c[0]-a[0])
}

var randPool = rand.New(rand.NewSource(rand.Int63())) // 重用随机数生成器

func (m *DelaunayMesh) locate(x [2]float64, startEdge *QuadEdge) *QuadEdge {
	if startEdge == nil {
		return m.startingQuadEdge // 返回默认起始边
	}
	e := startEdge
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
				break
			} else {
				t = to
				e = eo
			}
		} else {
			if to > 0 {
				if td == 0 && t == 0 {
					break
				} else {
					t = td
					e = ed
				}
			} else {
				if t == 0 && !leftOf(eo.Dest(), e) {
					e = e.Sym()
				} else if randPool.Intn(2) == 0 { // 优化随机数生成
					t = to
					e = eo
				} else {
					t = td
					e = ed
				}
			}
		}
	}

	m.startingQuadEdge = e
	return e
}

func (m *DelaunayMesh) spoke(x [2]float64, e *QuadEdge) *QuadEdge {
	var newFaces [4]*DelaunayTriangle
	faceCount := 0

	var boundaryEdge *QuadEdge

	// 处理左边面 (添加nil检查)
	leftFace := e.LeftFace()
	if leftFace != nil {
		leftFace.dontAnchor(e)
		newFaces[faceCount] = leftFace
		faceCount++
	}

	// 检查是否在边上
	onEdge := m.onQuadEdge(x, e)
	if onEdge {
		if m.ccwBoundary(e) {
			boundaryEdge = e
		} else {
			// 处理对称面 (添加nil检查)
			symFace := e.Sym().LeftFace()
			if symFace != nil {
				symFace.dontAnchor(e.Sym())
				newFaces[faceCount] = symFace
				faceCount++
			}

			// 删除多余的边
			e = e.OrigPrev()
			m.delete(e.OrigNext())
		}
	}

	// 创建新基础边
	base := New(m.QuadEdges)
	base.Init()
	base.SetEndPoints(e.Orig(), x)
	Splice(base, e)

	m.startingQuadEdge = base
	current := base
	for {
		current = m.connect(e, current.Sym())
		e = current.OrigPrev()
		if e.LeftNext() == m.startingQuadEdge {
			break
		}
	}

	// 删除边界边（如果需要）
	if boundaryEdge != nil {
		m.delete(boundaryEdge)
	}

	// 设置起始点
	if boundaryEdge != nil {
		current = m.startingQuadEdge.RightPrev()
	} else {
		current = m.startingQuadEdge.Sym()
	}

	// 重塑或创建新面
	for i := 0; i < faceCount; i++ {
		newFaces[i].reshape(current)
		current = current.OrigNext()
	}

	// 创建剩余的新面
	for current != m.startingQuadEdge.Sym() {
		m.makeFace(current)
		current = current.OrigNext()
	}

	return m.startingQuadEdge
}

func (m *DelaunayMesh) optimize(x [2]float64, s *QuadEdge) {
	queue := []*QuadEdge{} // 使用队列而不是栈
	visited := make(map[*QuadEdge]bool)

	// 初始时添加所有与点x相连的边
	start := s
	current := s
	for {
		// 只添加内部边
		if m.isInterior(current.LeftNext()) {
			if !visited[current.LeftNext()] {
				visited[current.LeftNext()] = true
				queue = append(queue, current.LeftNext())
			}
		}

		current = current.OrigNext()
		if current == start {
			break
		}
	}

	// 处理队列中的边
	for len(queue) > 0 {
		// 从队列前端取出
		e := queue[0]
		queue = queue[1:]

		// 确保边仍然存在且是内部边
		if e == nil || !m.isInterior(e) {
			continue
		}

		if m.shouldSwap(x, e) {
			// 保存受影响的邻居边
			affectedEdges := []*QuadEdge{
				e.OrigPrev(),
				e.Sym().OrigPrev(),
				e.OrigNext(),
				e.Sym().OrigNext(),
			}

			// 执行交换
			m.swap(e)

			// 将受影响的边加入队列
			for _, edge := range affectedEdges {
				if edge != nil && !visited[edge] && m.isInterior(edge) {
					visited[edge] = true
					queue = append(queue, edge)
				}
			}
		}
	}

	// 扫描所有三角形（可选）
	current = s
	for {
		if t := current.LeftFace(); t != nil && m.scanTriangle != nil {
			m.scanTriangle(t)
		}
		current = current.OrigNext()
		if current == s {
			break
		}
	}
}

func (m *DelaunayMesh) Insert(x [2]float64, tri *DelaunayTriangle) { // Capitalize method name
	var e *QuadEdge
	if tri != nil {
		e = m.locate(x, tri.Anchor)
	} else {
		e = m.locate(x, m.startingQuadEdge)
	}

	if isEqual(x, e.Orig()) || isEqual(x, e.Dest()) {
		m.optimize(x, e)
	} else {
		startSpoke := m.spoke(x, e)
		if startSpoke != nil {
			m.optimize(x, startSpoke.Sym())
		}
	}
}
