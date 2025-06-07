package tin

import (
	"math"
)

type Mesh struct {
	Vertices     []Vertex
	Normals      []Normal
	Faces        []Face
	Triangles    []Triangle
	BBox         [2][3]float64
	DiscardFaces []Face
}

func (m *Mesh) initFromDecomposed(vertices []Vertex, faces []Face, norl []Normal) {
	m.Vertices = vertices
	m.Faces = faces
	m.Normals = norl
	m.Triangles = make([]Triangle, 0)
}

func (m *Mesh) initFromTriangles(triangles []Triangle) {
	m.Triangles = triangles
	m.Faces = make([]Face, 0)
	m.Vertices = make([]Vertex, 0)
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

func (m *Mesh) addTriangle(t Triangle, decompose bool) {
	m.Triangles = append(m.Triangles, t)
	if decompose || m.hasDecomposed() {
		m.decomposeTriangle(t)
	}
}

func (m *Mesh) generateTriangles() {
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

func (m *Mesh) generateDecomposed() {
	if m.hasDecomposed() {
		return
	}

	m.Faces = make([]Face, len(m.Triangles))
	m.Vertices = make([]Vertex, 0, int(float64(len(m.Triangles))*0.66))

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
			}
		}
		m.Faces = append(m.Faces, f)
	}
}

func (m *Mesh) clearTriangles() {
	m.Triangles = make([]Triangle, 0)
}

func (m *Mesh) clearDecomposed() {
	m.Faces = make([]Face, 0)
	m.Vertices = make([]Vertex, 0)
}

func (m *Mesh) grabTriangles(into []Triangle) []Triangle {
	temp := m.Triangles
	m.Triangles = into
	return temp
}

func (m *Mesh) grabDecomposed(vertices []Vertex, faces []Face) ([]Vertex, []Face) {
	tempV := m.Vertices
	tempF := m.Faces
	m.Vertices = tempV
	m.Faces = tempF
	return tempV, tempF
}

func (m *Mesh) composeTriangle(f Face) *Triangle {
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
	for i := range m.Vertices {
		out.Add(m.Vertices[i])
	}
	return out
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
			}
		}
	}

	m.Faces = append(m.Faces, f)
}

func faceEdgeCrossesOtherEdge(fi int, faces []Face, vertices []Vertex) bool {
	f := faces[fi]
	faceEdges := []Edge{
		Edge{f[0], f[1]}, Edge{f[1], f[2]}, Edge{f[2], f[0]},
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

func adjacentFind(vi []Vertex, f2 func(Vertex, Vertex) bool) (index int, found bool) {
	boundOfVi := len(vi) - 1
	for i := 0; i < boundOfVi; i++ {
		if f2(vi[i], vi[i+1]) {
			return i, true
		}
	}
	return -1, false
}

func sort(v []Vertex, cp func(Vertex, Vertex) bool) {
	size := len(v)
	if size <= 1 {
		return
	}
	pivot := size / 2
	x := v[pivot]
	v[pivot], v[size-1] = v[size-1], v[pivot]
	j := 0
	for i := 0; i < size-1; i++ {
		if cp(v[i], x) {
			v[i], v[j] = v[j], v[i]
			j++
		}
	}
	v[j], v[size-1] = v[size-1], v[j]
	sort(v[:j], cp)
	sort(v[j+1:], cp)
}

func partialSortCopy(v []Vertex, out []Vertex, cp func(Vertex, Vertex) bool) {
	in := v
	inSize := len(in)
	outSize := len(out)
	if inSize <= outSize {
		copy(out, in)
		sort(out[:inSize], cp)
	} else {
		space := make([]Vertex, inSize)
		copy(space, in)
		sort(space, cp)
		copy(out, space)
	}
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

		partialSortCopy(m.Vertices, verticesSorted, VertexEqual)
		_, ok := adjacentFind(verticesSorted, VertexEqual)
		if ok {
			return false
		}
	}

	for fi := 0; fi < len(m.Faces); fi++ {
		if faceEdgeCrossesOtherEdge(fi, m.Faces, m.Vertices) {
			return false
		}
	}

	return true
}
