// Copyright (c) 2017-present FlyWave, Inc. All Rights Reserved.
// See License.txt for license information.

package tin

import "math"

func averageOf(d1, d2, d3, d4, noDataValue float64) float64 {
	count := 0
	sum := float64(0.0)
	lp := []float64{d1, d2, d3, d4}
	for d := range lp {
		if isNoData(lp[d], noDataValue) {
			continue
		}
		count++
		sum += lp[d]
	}

	if count > 0 {
		return sum / float64(count)
	}
	return math.NaN()
}

type ZemlyaMesh struct {
	Sample       RasterDouble
	Insert       RasterDouble
	Result       RasterDouble
	Used         RasterChar
	Token        RasterInt
	Candidates   CandidateList
	MaxError     float64
	Counter      int
	CurrentLevel int
	MaxLevel     int
}

func (z *ZemlyaMesh) scanTriangleLine(plane Plane, y int, x1, x2 float64, candidate Candidate, noDataValue float64) {

}

func (z *ZemlyaMesh) GreedyInsert(maxError float64) {

}

func (z *ZemlyaMesh) ScanTriangle() {

}
