package tin

import (
	"fmt"
	"math"
	"os"
	"sort"

	"github.com/flywave/go-geo"
)

type Mesh struct {
	GeoRef    *geo.GeoReference
	Vertices  []Vertex
	Normals   []Normal
	Faces     []Face
	Triangles []Triangle
	BBox      [2][3]float64
}

// 添加初始化方法
func NewMesh(geoRef *geo.GeoReference) *Mesh {
	m := &Mesh{
		GeoRef: geoRef,
	}
	m.initBBox()
	return m
}

// 更新包围盒，添加顶点
func (m *Mesh) updateBBox(v Vertex) {
	if m.BBox[0][0] > v[0] {
		m.BBox[0][0] = v[0]
	}
	if m.BBox[0][1] > v[1] {
		m.BBox[0][1] = v[1]
	}
	if m.BBox[0][2] > v[2] {
		m.BBox[0][2] = v[2]
	}

	if m.BBox[1][0] < v[0] {
		m.BBox[1][0] = v[0]
	}
	if m.BBox[1][1] < v[1] {
		m.BBox[1][1] = v[1]
	}
	if m.BBox[1][2] < v[2] {
		m.BBox[1][2] = v[2]
	}
}

// 初始化包围盒
func (m *Mesh) initBBox() {
	m.BBox[0] = [3]float64{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64}
	m.BBox[1] = [3]float64{-math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64}
}

func (m *Mesh) initFromDecomposed(vertices []Vertex, faces []Face, norl []Normal) {
	m.Vertices = vertices
	m.Faces = faces
	m.Normals = norl
	m.Triangles = make([]Triangle, 0)

	// 初始化并计算包围盒
	m.initBBox()
	for _, v := range vertices {
		m.updateBBox(v)
	}
}

func (m *Mesh) InitFromTriangles(triangles []Triangle) {
	m.Triangles = triangles
	m.Faces = make([]Face, 0)
	m.Vertices = make([]Vertex, 0)
	m.initBBox()
}

// 添加三角形有效性检查
func (t Triangle) IsValid() bool {
	// 检查三个顶点是否不同
	if t[0].Equal(t[1]) || t[0].Equal(t[2]) || t[1].Equal(t[2]) {
		return false
	}

	// 检查面积是否足够大
	v1 := [3]float64{t[1][0] - t[0][0], t[1][1] - t[0][1], t[1][2] - t[0][2]}
	v2 := [3]float64{t[2][0] - t[0][0], t[2][1] - t[0][1], t[2][2] - t[0][2]}

	// 计算叉积
	cross := [3]float64{
		v1[1]*v2[2] - v1[2]*v2[1],
		v1[2]*v2[0] - v1[0]*v2[2],
		v1[0]*v2[1] - v1[1]*v2[0],
	}

	// 计算面积
	area := math.Sqrt(cross[0]*cross[0] + cross[1]*cross[1] + cross[2]*cross[2])
	return area > EPS
}

// 优化添加三角形方法
func (m *Mesh) AddTriangle(t Triangle, decompose bool) {
	// 跳过无效三角形
	if !t.IsValid() {
		return
	}

	m.Triangles = append(m.Triangles, t)
	if decompose || m.hasDecomposed() {
		m.decomposeTriangle(t)
	}

	// 更新包围盒
	for _, v := range t {
		m.updateBBox(v)
	}
}

func (m *Mesh) decomposeTriangle(t Triangle) {
	f := Face{math.MaxUint32, math.MaxUint32, math.MaxUint32}
	found := 0

	for j := 0; j < len(m.Vertices) && found < 3; j++ {
		v := m.Vertices[j]
		for i := 0; i < 3; i++ {
			if t[i].Equal(v) {
				f[i] = VertexIndex(j)
				found++
			}
		}
	}

	if found != 3 {
		for i := 0; i < 3; i++ {
			if f[i] == math.MaxUint32 {
				f[i] = VertexIndex(len(m.Vertices))
				m.Vertices = append(m.Vertices, t[i])
				m.updateBBox(t[i]) // 更新包围盒
			}
		}
	}

	m.Faces = append(m.Faces, f)
}

func (m *Mesh) GenerateDecomposed() {
	if m.hasDecomposed() {
		return
	}

	m.Faces = make([]Face, 0, len(m.Triangles))
	m.Vertices = make([]Vertex, 0, int(float64(len(m.Triangles))*0.66))
	m.initBBox() // 初始化包围盒

	vertexLookup := make(map[Vertex]VertexIndex)

	for t := range m.Triangles {
		var f Face
		for i := 0; i < 3; i++ {
			v := m.Triangles[t][i]
			val, ok := vertexLookup[v]
			if ok {
				f[i] = val
			} else {
				f[i] = VertexIndex(len(m.Vertices))
				vertexLookup[v] = VertexIndex(len(m.Vertices))
				m.Vertices = append(m.Vertices, v)
				m.updateBBox(v) // 更新包围盒
			}
		}
		m.Faces = append(m.Faces, f)
	}
}

func (m *Mesh) Count() int {
	if m.hasTriangles() {
		return len(m.Triangles)
	}
	return len(m.Faces)
}

func (m *Mesh) Empty() bool {
	return m.Count() == 0
}

func (m *Mesh) hasTriangles() bool {
	return len(m.Triangles) != 0
}

func (m *Mesh) hasDecomposed() bool {
	return len(m.Vertices) != 0 && len(m.Faces) != 0
}

func (m *Mesh) GenerateTriangles() {
	if m.hasTriangles() {
		return
	}

	m.Triangles = make([]Triangle, len(m.Faces))
	var t Triangle
	for f := range m.Faces {
		isValid := true
		for i := 0; i < 3; i++ {
			vi := m.Faces[f][i]
			if int(vi) < len(m.Vertices) {
				t[i] = m.Vertices[vi]
			} else {
				isValid = false
			}
		}
		if isValid {
			m.Triangles[f] = t
		}
	}
}

func (m *Mesh) ClearTriangles() {
	m.Triangles = make([]Triangle, 0)
}

func (m *Mesh) ClearDecomposed() {
	m.Faces = make([]Face, 0)
	m.Vertices = make([]Vertex, 0)
}

func (m *Mesh) GrabTriangles(into []Triangle) []Triangle {
	temp := m.Triangles
	m.Triangles = into
	return temp
}

func (m *Mesh) GrabDecomposed(vertices []Vertex, faces []Face) ([]Vertex, []Face) {
	tempV := m.Vertices
	tempF := m.Faces
	m.Vertices = tempV
	m.Faces = tempF
	return tempV, tempF
}

func (m *Mesh) ComposeTriangle(f Face) *Triangle {
	out := new(Triangle)
	for i := 0; i < 3; i++ {
		if int(f[i]) >= len(m.Vertices) {
			return nil
		}
	}
	for i := 0; i < 3; i++ {
		out[i] = m.Vertices[f[i]]
	}
	return out
}

func (m *Mesh) GetBbox() BBox3d {
	var out BBox3d
	out[0], out[1], out[2] = m.BBox[0][1], m.BBox[0][1], m.BBox[0][2]
	out[3], out[4], out[5] = m.BBox[1][0], m.BBox[1][1], m.BBox[1][2]
	return out
}

func faceEdgeCrossesOtherEdge(fi int, faces []Face, vertices []Vertex) bool {
	f := faces[fi]
	faceEdges := []Edge{
		{f[0], f[1]}, {f[1], f[2]}, {f[2], f[0]},
	}

	var ft Triangle
	ft[0] = vertices[f[0]]
	ft[1] = vertices[f[1]]
	ft[2] = vertices[f[2]]

	ftbbox := NewBBox2d()
	ftbbox.Add(ft)

	for oi := fi + 1; oi < len(faces); oi++ {
		o := faces[oi]

		ot := Triangle{
			vertices[o[0]], vertices[o[1]], vertices[o[2]],
		}

		otbbox := NewBBox2d()
		otbbox.Add(ot)

		if ftbbox.Intersects(*otbbox, EPS) {
			for i := range faceEdges {
				e := faceEdges[i]
				oe := Edge{o[0], o[1]}
				if !e.SharesPoint(oe) && e.Intersects2d(oe, vertices) {
					return true
				}

				oe.Assign(o[1], o[2])
				if !e.SharesPoint(oe) && e.Intersects2d(oe, vertices) {
					return true
				}
				oe.Assign(o[2], o[0])
				if !e.SharesPoint(oe) && e.Intersects2d(oe, vertices) {
					return true
				}
			}
		}
	}

	return false
}

// 优化排序函数
func sortVertices(v []Vertex, less func(Vertex, Vertex) bool) {
	sort.Slice(v, func(i, j int) bool {
		return less(v[i], v[j])
	})
}

func partialSortCopy(v []Vertex, out []Vertex, less func(Vertex, Vertex) bool) {
	// 先复制数据
	copy(out, v)

	// 对副本进行部分排序
	if len(out) > len(v) {
		sortVertices(out[:len(v)], less)
	} else {
		sortVertices(out, less)
	}
}

// 添加辅助函数用于顶点比较
func vertexLess(a, b Vertex) bool {
	if a[0] != b[0] {
		return a[0] < b[0]
	}
	if a[1] != b[1] {
		return a[1] < b[1]
	}
	return a[2] < b[2]
}

func (m *Mesh) CheckTin() bool {
	if !m.hasDecomposed() {
		return false
	}

	{
		verticesSize := len(m.Vertices)
		vertexUsed := make([]bool, verticesSize)
		for i := range vertexUsed {
			vertexUsed[i] = false
		}
		for f := range m.Faces {
			if int(m.Faces[f][0]) >= verticesSize || int(m.Faces[f][1]) >= verticesSize ||
				int(m.Faces[f][2]) >= verticesSize {
				return false
			}

			vertexUsed[m.Faces[f][0]] = true
			vertexUsed[m.Faces[f][1]] = true
			vertexUsed[m.Faces[f][2]] = true

			if m.Faces[f][0] == m.Faces[f][1] || m.Faces[f][0] == m.Faces[f][2] || m.Faces[f][1] == m.Faces[f][2] {
				return false
			}

			if !isFacingUpwardsForFace(m.Faces[f], m.Vertices) {
				return false
			}
		}

		for i := range vertexUsed {
			if !vertexUsed[i] {
				return false
			}
		}
	}

	{
		verticesSorted := make([]Vertex, len(m.Vertices))
		partialSortCopy(m.Vertices, verticesSorted, vertexLess)

		// 使用更高效的检查重复顶点方式
		for i := 0; i < len(verticesSorted)-1; i++ {
			if verticesSorted[i].Equal(verticesSorted[i+1]) {
				return false
			}
		}
	}

	for fi := 0; fi < len(m.Faces); fi++ {
		if faceEdgeCrossesOtherEdge(fi, m.Faces, m.Vertices) {
			return false
		}
	}

	return true
}

// 新增OBJ导出方法
func (m *Mesh) ExportOBJ(filename string, reproj bool) error {
	srcProj := geo.NewProj(3857) // 当前坐标系
	dstProj := m.GeoRef.GetSrs() // 目标坐标系（原始坐标系）
	convert := func(x, y float64) (float64, float64) {
		pt, _ := transformPoint(srcProj, dstProj, x, y)
		return pt[0], pt[1]
	}
	// 创建输出文件
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	// 写入顶点数据
	file.WriteString("# Vertices\n")
	for _, v := range m.Vertices {
		if reproj {
			x, y := convert(v[0], v[1])
			file.WriteString(fmt.Sprintf("v %.6f %.6f %.6f\n", x, y, v[2]))
		} else {
			file.WriteString(fmt.Sprintf("v %.6f %.6f %.6f\n", v[0], v[1], v[2]))
		}
	}

	// 写入法线数据（如果存在）
	if len(m.Normals) > 0 {
		file.WriteString("\n# Normals\n")
		for _, n := range m.Normals {
			file.WriteString(fmt.Sprintf("vn %.6f %.6f %.6f\n", n[0], n[1], n[2]))
		}
	}

	// 写入面数据
	file.WriteString("\n# Faces\n")
	for _, f := range m.Faces {
		if len(m.Normals) > 0 {
			// 顶点索引和法线索引都从1开始
			file.WriteString(fmt.Sprintf("f %d//%d %d//%d %d//%d\n",
				f[0]+1, f[0]+1,
				f[1]+1, f[1]+1,
				f[2]+1, f[2]+1))
		} else {
			file.WriteString(fmt.Sprintf("f %d %d %d\n",
				f[0]+1, f[1]+1, f[2]+1))
		}
	}

	return nil
}
