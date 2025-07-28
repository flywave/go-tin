package tin

import "math"

func Orientation(a, b, c [2]float64) float64 {
	ab0 := b[0] - a[0]
	ab1 := b[1] - a[1]
	ac0 := c[0] - a[0]
	ac1 := c[1] - a[1]
	return ab0*ac1 - ab1*ac0
}

func IsCCW(a, b, c [2]float64) bool {
	return Orientation(a, b, c) > 0
}

func InTriangle(a, b, c, p [2]float64) bool {
	// 要求三角形点序为逆时针 (CCW)
	ab0, ab1 := b[0]-a[0], b[1]-a[1]
	ac0, ac1 := c[0]-a[0], c[1]-a[1]
	ap0, ap1 := p[0]-a[0], p[1]-a[1]

	// 计算叉积 (重心坐标的分子)
	u := ac1*ap0 - ac0*ap1
	if u < 0 {
		return false
	}
	v := -ab1*ap0 + ab0*ap1
	if v < 0 {
		return false
	}

	// 三角形面积的2倍 (分母)
	area2 := ab0*ac1 - ab1*ac0
	// 添加对退化三角形的检查
	if area2 <= 0 {
		return false
	}
	return u+v <= area2
}

func InCircumcircle(a, b, c, d [2]float64) bool {
	// 检查是否有重复点
	if isEqual(a, b) || isEqual(a, c) || isEqual(a, d) ||
		isEqual(b, c) || isEqual(b, d) || isEqual(c, d) {
		return false
	}

	// 计算向量
	ab := sub(b, a)
	ac := sub(c, a)

	// 检查三角形退化
	cross := ab[0]*ac[1] - ab[1]*ac[0]
	if math.Abs(cross) < EPS {
		return false // 退化三角形，不处理
	}

	// 使用高精度行列式计算
	// 创建矩阵：
	// | ax-dx, ay-dy, (ax-dx)^2 + (ay-dy)^2 |
	// | bx-dx, by-dy, (bx-dx)^2 + (by-dy)^2 |
	// | cx-dx, cy-dy, (cx-dx)^2 + (cy-dy)^2 |
	adx := a[0] - d[0]
	ady := a[1] - d[1]
	bdx := b[0] - d[0]
	bdy := b[1] - d[1]
	cdx := c[0] - d[0]
	cdy := c[1] - d[1]

	// 计算平方距离
	ad2 := adx*adx + ady*ady
	bd2 := bdx*bdx + bdy*bdy
	cd2 := cdx*cdx + cdy*cdy

	// 计算行列式
	det := adx*(bdy*cd2-bd2*cdy) -
		ady*(bdx*cd2-bd2*cdx) +
		ad2*(bdx*cdy-bdy*cdx)

	// 计算三角形三条边的平方长度
	abLenSq := ab[0]*ab[0] + ab[1]*ab[1]
	acLenSq := ac[0]*ac[0] + ac[1]*ac[1]
	bcLenSq := (c[0]-b[0])*(c[0]-b[0]) + (c[1]-b[1])*(c[1]-b[1])
	maxEdgeSq := math.Max(abLenSq, math.Max(acLenSq, bcLenSq))

	// 计算动态误差阈值，使用较小系数确保阈值足够小
	eps := math.Max(1e-12, maxEdgeSq*1e-8)

	// 判断点是否在圆内
	return det > eps
}

func Circumcenter(a, b, c [2]float64) [2]float64 {
	// 计算向量 BA 和 CA
	ba0, ba1 := b[0]-a[0], b[1]-a[1]
	ca0, ca1 := c[0]-a[0], c[1]-a[1]

	// 计算叉积（三角形面积的2倍）
	cross := ba0*ca1 - ba1*ca0

	// 处理退化情况（三点共线）
	if math.Abs(cross) < 1e-15 {
		return [2]float64{math.NaN(), math.NaN()}
	}

	// 预计算长度平方
	lba := ba0*ba0 + ba1*ba1
	lca := ca0*ca0 + ca1*ca1

	// 计算逆矩阵系数
	d := 0.5 / cross

	// 计算外心坐标（相对于点A）
	x := (ca1*lba - ba1*lca) * d
	y := (ba0*lca - ca0*lba) * d

	// 转换为绝对坐标
	return [2]float64{x + a[0], y + a[1]}
}
