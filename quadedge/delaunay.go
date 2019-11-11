package quadedge

import (
	"sort"
)

func Delaunay(points [][2]float64) Edge {
	v := make([]uint32, 0, len(points))
	for i := range points {
		n := true
		for j := range points[:i] {
			if points[j] == points[i] {
				n = false
				break
			}
		}
		if n {
			v = append(v, uint32(i))
		}
	}
	if len(v) < 2 {
		return Edge{id: Nil}
	}

	sort.Slice(v, func(i int, j int) bool {
		switch {
		case points[v[i]][0] < points[v[j]][0]:
			return true
		case points[v[i]][0] > points[v[j]][0]:
			return false
		default:
			return points[v[i]][1] < points[v[j]][1]
		}
	})

	p := NewPool(uint32(len(v) * 3))
	l, _ := delaunay(points, p, v)

	return l
}

func delaunay(points [][2]float64, p *Pool, sub []uint32) (l, r Edge) {
	if len(sub) == 2 {
		a := New(p)
		a.SetOrig(sub[0])
		a.SetDest(sub[1])
		return a, a.Sym()
	}

	if len(sub) == 3 {
		a := New(p)
		b := New(p)
		Splice(a.Sym(), b)
		a.SetOrig(sub[0])
		a.SetDest(sub[1])
		b.SetOrig(sub[1])
		b.SetDest(sub[2])

		if IsCCW(points[sub[0]], points[sub[1]], points[sub[2]]) {
			_ = Connect(b, a)
			return a, b.Sym()
		}
		if IsCCW(points[sub[0]], points[sub[2]], points[sub[1]]) {
			c := Connect(b, a)
			return c.Sym(), c
		}

		return a, b.Sym()
	}

	lout, lins := delaunay(points, p, sub[:len(sub)/2])
	rins, rout := delaunay(points, p, sub[len(sub)/2:])

loop:
	for {
		switch {
		case leftOf(points, rins.Orig(), lins):
			lins = lins.LeftNext()
		case rightOf(points, lins.Orig(), rins):
			rins = rins.RightPrev()
		default:
			break loop
		}
	}

	base := Connect(rins.Sym(), lins)
	if lins.Orig() == lout.Orig() {
		lout = base.Sym()
	}
	if rins.Orig() == rout.Orig() {
		rout = base
	}

	for {
		lcand := base.Sym().OrigNext()
		if valid(points, lcand, base) {
			for inCircle(points,
				base.Dest(), base.Orig(), lcand.Dest(), lcand.OrigNext().Dest(),
			) {
				t := lcand.OrigNext()
				Delete(lcand)
				lcand = t
			}
		}

		rcand := base.OrigPrev()
		if valid(points, rcand, base) {
			for inCircle(points,
				base.Dest(), base.Orig(), rcand.Dest(), rcand.OrigPrev().Dest(),
			) {
				t := rcand.OrigPrev()
				Delete(rcand)
				rcand = t
			}
		}

		if !valid(points, lcand, base) && !valid(points, rcand, base) {
			break
		}

		if !valid(points, lcand, base) ||
			(valid(points, rcand, base) &&
				inCircle(points,
					lcand.Dest(), lcand.Orig(), rcand.Orig(), rcand.Dest())) {
			base = Connect(rcand, base.Sym())
		} else {
			base = Connect(base.Sym(), lcand.Sym())
		}
	}

	return lout, rout
}

func inCircle(points [][2]float64, a, b, c, d uint32) bool {
	return InCircumcircle(points[a], points[b], points[c], points[d])
}

func rightOf(points [][2]float64, p uint32, e Edge) bool {
	return IsCCW(points[p], points[e.Dest()], points[e.Orig()])
}

func leftOf(points [][2]float64, p uint32, e Edge) bool {
	return IsCCW(points[p], points[e.Orig()], points[e.Dest()])
}

func valid(points [][2]float64, e, f Edge) bool {
	return rightOf(points, e.Dest(), f)
}
