// Copyright (c) 2017-present FlyWave, Inc. All Rights Reserved.
// See License.txt for license information.

package tin

type Candidate struct {
}

type CandidateList struct {
}

func isNoData(d, noDataValue float64) bool {
	return false
}

type RasterMesh struct {
	DelaunayMesh
	Raster RasterDouble
}

func (r *RasterMesh) repairPoint(px, py float64) {

}
