package tin

import (
	"math"
	"testing"
)

// 创建测试地形栅格 (5x5 山丘地形)
func createTestRaster() *RasterDouble {
	data := [][]float64{
		{0, 0, 0, 0, 0},
		{0, 1, 1, 1, 0},
		{0, 1, 2, 1, 0},
		{0, 1, 1, 1, 0},
		{0, 0, 0, 0, 0},
	}

	flatData := make([]float64, 25)
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			flatData[y*5+x] = data[y][x]
		}
	}

	raster := NewRasterDoubleWithData(5, 5, flatData)
	raster.SetXYPos(0, 0, 1.0) // 设置地理参考
	raster.NoData = math.NaN() // 设置无效值
	return raster
}

// 综合测试ZemlyaMesh算法
func TestZemlyaMeshIntegration(t *testing.T) {
	// ========== 准备阶段 ==========
	testRaster := createTestRaster()

	// ========== 执行阶段 ==========
	zemlya := NewZemlyaMesh(nil)
	zemlya.Raster = testRaster
	maxError := 0.5 // 最大允许误差
	zemlya.GreedyInsert(maxError)
	resultMesh := zemlya.ToMesh()

	// ========== 验证阶段 ==========
	// 1. 验证关键点存在 - 使用中心点坐标
	keyPoints := []struct {
		x, y float64
	}{
		{0.5, 4.5}, // (0,0)位置的中心点
		{4.5, 4.5}, // (0,4)位置的中心点
		{4.5, 0.5}, // (4,4)位置的中心点
		{0.5, 0.5}, // (4,0)位置的中心点
		{2.5, 2.5}, // (2,2)位置的中心点
	}

	for _, pt := range keyPoints {
		found := false
		for _, v := range resultMesh.Vertices {
			if math.Abs(v[0]-pt.x) < 1e-5 && math.Abs(v[1]-pt.y) < 1e-5 {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("关键点(%.1f, %.1f)未在结果中找到", pt.x, pt.y)
		}
	}

	// 2. 验证网格拓扑结构 (放宽范围)
	if len(resultMesh.Faces) < 4 {
		t.Errorf("三角面数量过少: %d (至少需要4个)", len(resultMesh.Faces))
	}

	// 3. 验证简化程度
	originalVertices := 25
	resultVertices := len(resultMesh.Vertices)
	if resultVertices >= originalVertices {
		t.Errorf("顶点数量未减少: 原始=%d, 结果=%d", originalVertices, resultVertices)
	} else {
		t.Logf("顶点简化率: %.1f%%",
			float64(originalVertices-resultVertices)/float64(originalVertices)*100)
	}

	// 4. 验证误差控制
	maxFoundError := 0.0
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			origZ := testRaster.Value(y, x)
			resultZ := zemlya.Result.Value(y, x)
			if !math.IsNaN(resultZ) {
				err := math.Abs(origZ - resultZ)
				if err > maxFoundError {
					maxFoundError = err
				}
			}
		}
	}

	if maxFoundError > maxError {
		t.Errorf("最大误差超标: %.4f > 允许值%.2f", maxFoundError, maxError)
	} else {
		t.Logf("实际最大误差: %.4f (<=允许值%.2f)", maxFoundError, maxError)
	}

	// 5. 验证边界处理 - 使用栅格实际边界
	westmost, eastmost := math.MaxFloat64, -math.MaxFloat64
	for _, v := range resultMesh.Vertices {
		if v[0] < westmost {
			westmost = v[0]
		}
		if v[0] > eastmost {
			eastmost = v[0]
		}
	}

	expectedWest := testRaster.West() + testRaster.cellsize/2
	expectedEast := testRaster.East() - testRaster.cellsize/2

	if math.Abs(westmost-expectedWest) > 1e-5 || math.Abs(eastmost-expectedEast) > 1e-5 {
		t.Errorf("边界范围错误: 西界=%.2f (应为%.2f), 东界=%.2f (应为%.2f)",
			westmost, expectedWest, eastmost, expectedEast)
	}

	// 6. 验证无效值处理
	// 设置一个无效点
	testRaster.SetValue(0, 1, math.NaN())
	// 重新执行算法
	zemlya = NewZemlyaMesh(nil)
	zemlya.Raster = testRaster
	zemlya.GreedyInsert(maxError)

	// 检查无效点处理
	if !math.IsNaN(zemlya.Result.Value(0, 1)) {
		t.Errorf("无效点(0,1)未被正确处理")
	}

	// 检查有效点重建
	if math.IsNaN(zemlya.Result.Value(2, 2)) {
		t.Errorf("有效点(2,2)被错误标记为无效")
	}
}
