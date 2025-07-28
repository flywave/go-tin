package tin

func GenerateTinMesh(raster *RasterDouble, maxError float64) *Mesh {
	g := NewZemlyaMesh()
	g.LoadRaster(raster)
	g.GreedyInsert(maxError)
	return g.ToMesh()
}

type TileMaker struct {
	mesh *Mesh
}

func NewTileMaker(m *Mesh) *TileMaker {
	return &TileMaker{mesh: m}
}

func CheckTriangleInTile(t Triangle, tileBounds BBox2d) bool {
	triangleBounds := NewBBox2d()
	tileBounds.Add(t)
	return triangleBounds.Intersects(tileBounds, EPS)
}

func (tm *TileMaker) GenTile(scale bool) (*Mesh, error) {
	vertsInTile := make([]Vertex, len(tm.mesh.Vertices))
	copy(vertsInTile, tm.mesh.Vertices)
	tileBBox := *NewBBox3d()
	tileBBox[0] = tm.mesh.BBox[0][0]
	tileBBox[1] = tm.mesh.BBox[0][1]
	tileBBox[2] = tm.mesh.BBox[0][2]

	tileBBox[3] = tm.mesh.BBox[1][0]
	tileBBox[4] = tm.mesh.BBox[1][1]
	tileBBox[5] = tm.mesh.BBox[1][2]
	if scale {
		tileInverseScaleX := 1.0
		tileInverseScaleY := 1.0
		tileInverseScaleZ := 1.0

		if width := tileBBox.Width(); width > 0 {
			tileInverseScaleX = 1.0 / width
		}
		if height := tileBBox.Height(); height > 0 {
			tileInverseScaleY = 1.0 / height
		}
		if depth := tileBBox.Depth(); depth > 0 {
			tileInverseScaleZ = 1.0 / depth
		}

		ctx := tileBBox[0]
		cty := tileBBox[1]
		ctz := tileBBox[2]

		for t := range vertsInTile {
			v := &vertsInTile[t]
			v[0] = (v[0] - ctx) * tileInverseScaleX
			v[1] = (v[1] - cty) * tileInverseScaleY
			v[2] = (v[2] - ctz) * tileInverseScaleZ
		}
	}

	fInTile := make([]Face, len(tm.mesh.Faces))
	copy(fInTile, tm.mesh.Faces)
	tileMesh := new(Mesh)
	tileMesh.initFromDecomposed(vertsInTile, fInTile, tm.mesh.Normals)

	return tileMesh, nil
}
