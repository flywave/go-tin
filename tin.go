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

func (tm *TileMaker) GenTile(tsf [6]float64, x, y int) (*Mesh, error) {
	vertsInTile := make([]Vertex, len(tm.mesh.Vertices))
	copy(vertsInTile, tm.mesh.Vertices)
	tileBBox := *NewBBox3d()
	tileBBox[0] = tsf[0]
	tileBBox[1] = tsf[3] + tsf[5]*float64(y)
	tileBBox[3] = tsf[0] + tsf[1]*float64(x)
	tileBBox[4] = tsf[3]
	ctx := (tileBBox[0] + tileBBox[3]) / 2
	cty := (tileBBox[1] + tileBBox[4]) / 2

	for t := range vertsInTile {
		v := &vertsInTile[t]
		if v[2] < tileBBox[2] {
			tileBBox[2] = v[2]
		}
		if v[2] > tileBBox[5] {
			tileBBox[5] = v[2]
		}
		v[0] = (v[0] - ctx)
		v[1] = (v[1] - cty)
	}

	ctz := (tileBBox[2] + tileBBox[5]) / 2
	for t := range vertsInTile {
		v := &vertsInTile[t]
		v[2] = (v[2] - ctx)
	}

	fInTile := make([]Face, len(tm.mesh.Faces))
	copy(fInTile, tm.mesh.Faces)
	tileMesh := new(Mesh)
	tileMesh.initFromDecomposed(vertsInTile, fInTile)
	tileMesh.BBox[0][0] = tileBBox[0] - ctx
	tileMesh.BBox[0][1] = tileBBox[1] - cty
	tileMesh.BBox[0][2] = tileBBox[2] - ctz
	tileMesh.BBox[1][0] = tileBBox[3] - ctx
	tileMesh.BBox[1][1] = tileBBox[4] - cty
	tileMesh.BBox[1][2] = tileBBox[5] - ctz

	return tileMesh, nil
}
