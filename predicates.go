package tin

func Orientation(a, b, c [2]float64) float32 {
	d := a[0]*b[1] + b[0]*c[1] + c[0]*a[1] - b[1]*c[0] - c[1]*a[0] - a[1]*b[0]
	return float32(d)
}

func IsCCW(a, b, c [2]float64) bool {
	d := (b[0]-a[0])*(c[1]-a[1]) - (b[1]-a[1])*(c[0]-a[0])
	return d > 0
}

func InTriangle(a, b, c [2]float64, p [2]float64) bool {
	s := a[1]*c[0] - a[0]*c[1] + (c[1]-a[1])*p[0] + (a[0]-c[0])*p[1]
	d := a[0]*b[1] - a[1]*b[0] + (a[1]-b[1])*p[0] + (b[0]-a[0])*p[1]

	if (s < 0) != (d < 0) {
		return false
	}

	r := -b[1]*c[0] + a[1]*(c[0]-b[0]) + a[0]*(b[1]-c[1]) + b[0]*c[1]

	if r < 0 {
		s = -s
		d = -d
		r = -r
	}
	return s > 0 && d > 0 && (s+d) <= r
}

func Minus(a, b [2]float64) [2]float64 {
	return [2]float64{a[0] - b[0], a[1] - b[1]}
}

func InTriangleCCW(a, b, c [2]float64, p [2]float64) bool {
	bb := Minus(b, a)
	cc := Minus(c, a)
	pp := Minus(p, a)

	w := [2]float64{
		cc[1]*pp[0] - cc[0]*pp[1],
		-bb[1]*pp[0] + bb[0]*pp[1],
	}
	if w[0] <= 0 || w[1] <= 0 {
		return false
	}
	d := bb[0]*cc[1] - cc[0]*bb[1]
	return w[0]+w[1] < d
}

func InCircumcircle(a, b, c, d [2]float64) bool {
    // 直接使用坐标值
    ax, ay := a[0], a[1]
    bx, by := b[0], b[1]
    cx, cy := c[0], c[1]
    dx, dy := d[0], d[1]
    
    // 计算行列式（完全展开）
    det := (bx*by - by*bx)*(cx*cx + cy*cy - ax*ax - ay*ay) +
           (cx*cy - cy*cx)*(dx*dx + dy*dy - ax*ax - ay*ay) +
           (dx*dy - dy*dx)*(bx*bx + by*by - ax*ax - ay*ay)
    
    return det > EPS
}

func Circumcenter(a, b, c [2]float64) [2]float64 {
	ba := Minus(b, a)
	ca := Minus(c, a)

	lba := ba[0]*ba[0] + ba[1]*ba[1]
	lca := ca[0]*ca[0] + ca[1]*ca[1]

	d := 0.5 / (ba[0]*ca[1] - ba[1]*ca[0])

	o := [2]float64{
		(ca[1]*lba - ba[1]*lca) * d,
		(ba[0]*lca - ca[0]*lba) * d,
	}

	return [2]float64{o[0] + a[0], o[1] + a[1]}
}
