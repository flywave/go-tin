// Copyright (c) 2017-present FlyWave, Inc. All Rights Reserved.
// See License.txt for license information.

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

func (tm *TileMaker) GenTile(tx, ty int64, zoom uint64) (*Mesh, error) {
	vertsInTile := make([]Vertex, len(tm.mesh.Vertices))
	copy(vertsInTile, tm.mesh.Vertices)

	tileBounds := TileBounds(tx, ty, zoom)
	tileOrigin := [2]float64{tileBounds[0], tileBounds[1]}
	tileBBox := *NewBBox3d()
	tileBBox[0] = tileBounds[0]
	tileBBox[1] = tileBounds[1]
	tileBBox[3] = tileBounds[2]
	tileBBox[4] = tileBounds[3]

	for t := range vertsInTile {
		v := &vertsInTile[t]
		if v[2] < tileBBox[2] {
			tileBBox[2] = v[2]
		}
		if v[2] > tileBBox[5] {
			tileBBox[5] = v[2]
		}
		v[0] = (v[0] - tileOrigin[0])
		v[1] = (v[1] - tileOrigin[1])
	}

	fInTile := make([]Face, len(tm.mesh.Faces))
	copy(fInTile, tm.mesh.Faces)
	tileMesh := new(Mesh)
	tileMesh.initFromDecomposed(vertsInTile, fInTile)
	tileMesh.BBox[0][0] = tileBBox[0]
	tileMesh.BBox[0][1] = tileBBox[1]
	tileMesh.BBox[0][2] = tileBBox[2]
	tileMesh.BBox[1][0] = tileBBox[3]
	tileMesh.BBox[1][1] = tileBBox[4]
	tileMesh.BBox[1][2] = tileBBox[5]

	return tileMesh, nil
}
