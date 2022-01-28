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

func checkTriangleInTile(t Triangle, tileBounds BBox2d) bool {
	triangleBounds := NewBBox2d()
	tileBounds.Add(t)
	return triangleBounds.Intersects(tileBounds, EPS)
}

func (tm *TileMaker) GenTile() (*Mesh, error) {
	vertsInTile := make([]Vertex, len(tm.mesh.Vertices))
	copy(vertsInTile, tm.mesh.Vertices)
	tileBBox := *NewBBox3d()
	tileBBox[0] = tm.mesh.BBox[0][0]
	tileBBox[1] = tm.mesh.BBox[0][1]
	tileBBox[2] = tm.mesh.BBox[0][2]

	tileBBox[3] = tm.mesh.BBox[1][0]
	tileBBox[4] = tm.mesh.BBox[1][1]
	tileBBox[5] = tm.mesh.BBox[1][2]

	fInTile := make([]Face, len(tm.mesh.Faces))
	copy(fInTile, tm.mesh.Faces)
	tileMesh := new(Mesh)
	tileMesh.initFromDecomposed(vertsInTile, fInTile, tm.mesh.Normals)
	tileMesh.BBox[0][0] = tileBBox[0]
	tileMesh.BBox[0][1] = tileBBox[1]
	tileMesh.BBox[0][2] = tileBBox[2]
	tileMesh.BBox[1][0] = tileBBox[3]
	tileMesh.BBox[1][1] = tileBBox[4]
	tileMesh.BBox[1][2] = tileBBox[5]

	return tileMesh, nil
}
