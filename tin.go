// Copyright (c) 2017-present FlyWave, Inc. All Rights Reserved.
// See License.txt for license information.

package tin

import (
	"errors"
)

func GenerateTinMesh(raster *RasterDouble, maxError float64) *Mesh {
	g := new(ZemlyaMesh)
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
	var vertsInTile []Vertex
	copy(vertsInTile, tm.mesh.Vertices)

	tileBounds := TileBounds(tx, ty, zoom)
	buffer := tileBounds.Width() / 4.0

	tileBoundsWithBuffer := BBox2d{tileBounds[0] - buffer, tileBounds[1] - buffer,
		tileBounds[2] + buffer, tileBounds[3] + buffer}

	var trianglesInTile []Triangle
	tm.mesh.generateTriangles()

	for i := range tm.mesh.Triangles {
		tp := tm.mesh.Triangles[i]
		if checkTriangleInTile(tp, tileBoundsWithBuffer) {
			trianglesInTile = append(trianglesInTile, tp)
		}
	}

	tileOrigin := [2]float64{tileBounds[0], tileBounds[1]}

	tileBBox := *NewBBox3d()

	tileBBox[0] = tileBounds[0]
	tileBBox[1] = tileBounds[1]
	tileBBox[3] = tileBounds[2]
	tileBBox[4] = tileBounds[3]

	for t := range trianglesInTile {
		triangle := trianglesInTile[t]
		for i := 0; i < 3; i++ {
			if triangle[i][2] < tileBBox[2] {
				tileBBox[2] = triangle[i][2]
			}
			if triangle[i][2] > tileBBox[5] {
				tileBBox[5] = triangle[i][2]
			}
		}
	}

	tileInverseScaleX := 1.0 / tileBounds.Width()
	tileInverseScaleY := 1.0 / tileBounds.Height()

	tileInverseScaleZ := 1.0 / (tileBBox[5] - tileBBox[2])

	for t := range trianglesInTile {
		triangle := trianglesInTile[t]
		for i := 0; i < 3; i++ {
			triangle[i][0] = (triangle[i][0] - tileOrigin[0]) * tileInverseScaleX
			triangle[i][1] = (triangle[i][1] - tileOrigin[1]) * tileInverseScaleY
			triangle[i][2] = (triangle[i][2] - tileBBox[2]) * tileInverseScaleZ
		}
	}

	trianglesInTile = Clip25dTrianglesTo01Quadrant(trianglesInTile)

	if len(vertsInTile) == 0 {
		return nil, errors.New("verts is zero")
	}

	var fInTile []Face

	copy(fInTile, tm.mesh.Faces)

	tileMesh := new(Mesh)
	tileMesh.initFromDecomposed(vertsInTile, fInTile)
	return tileMesh, nil
}
