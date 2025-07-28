package tin

import (
	"math"
	"testing"
)

// 测试不同数据类型的栅格创建和基本操作
func TestRasterCreationAndBasicOperations(t *testing.T) {
	tests := []struct {
		name     string
		dataType int
		noData   interface{}
		setValue interface{}
	}{
		{"Int8", RASTER_DATA_TYPE_INT8, int8(-10), int8(42)},
		{"Uint8", RASTER_DATA_TYPE_UINT8, uint8(0), uint8(255)},
		{"Int16", RASTER_DATA_TYPE_INT16, int16(-1000), int16(32000)},
		{"Uint16", RASTER_DATA_TYPE_UINT16, uint16(0), uint16(65535)},
		{"Int32", RASTER_DATA_TYPE_INT32, int32(-100000), int32(2147483647)},
		{"Uint32", RASTER_DATA_TYPE_UINT32, uint32(0), uint32(4294967295)},
		{"Float32", RASTER_DATA_TYPE_FLOAT32, float32(math.NaN()), float32(3.14159)},
		{"Float64", RASTER_DATA_TYPE_FLOAT64, math.NaN(), 2.71828},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试 NewRaster
			r := NewRaster(5, 5, tt.dataType)
			if r.Rows() != 5 || r.Cols() != 5 {
				t.Errorf("NewRaster 尺寸错误: 预期 5x5, 实际 %dx%d", r.Rows(), r.Cols())
			}

			// 测试 SetValue 和 Value
			r.SetValue(2, 2, tt.setValue)
			val := r.Value(2, 2)
			if val != tt.setValue {
				t.Errorf("值设置/读取错误: 预期 %v, 实际 %v", tt.setValue, val)
			}

			// 测试 GetRow
			row := r.GetRow(2)
			if row == nil {
				t.Error("GetRow 返回 nil")
			}

			// 测试无效值处理
			r.SetValue(3, 3, tt.noData)
			val = r.Value(3, 3)
			if IsNaN(tt.noData) {
				if !IsNaN(val) {
					t.Errorf("无效值处理错误: 预期 NaN, 实际 %v", val)
				}
			} else if val != tt.noData {
				t.Errorf("无效值处理错误: 预期 %v, 实际 %v", tt.noData, val)
			}

			// 测试 NewRasterWithNoData
			r2 := NewRasterWithNoData(3, 3, tt.noData)
			if r2.Rows() != 3 || r2.Cols() != 3 {
				t.Errorf("NewRasterWithNoData 尺寸错误: 预期 3x3, 实际 %dx%d", r2.Rows(), r2.Cols())
			}
			for i := 0; i < 3; i++ {
				for j := 0; j < 3; j++ {
					val := r2.Value(i, j)
					if IsNaN(tt.noData) {
						if !IsNaN(val) {
							t.Errorf("NewRasterWithNoData 初始化错误: 位置 (%d,%d) 的值不是 NaN", i, j)
						}
					} else if val != tt.noData {
						t.Errorf("NewRasterWithNoData 初始化错误: 位置 (%d,%d) 的值不是 %v", i, j, tt.noData)
					}
				}
			}
		})
	}
}

// 测试栅格地理坐标转换
func TestRasterCoordinateTransformation(t *testing.T) {
	// 创建栅格并设置位置和分辨率
	r := NewRaster(10, 20, RASTER_DATA_TYPE_FLOAT64)
	r.SetXYPos(100.0, 200.0, 5.0) // 西:100, 南:200, 分辨率:5

	// 验证边界
	if r.West() != 100.0 || r.South() != 200.0 {
		t.Errorf("边界设置错误: 西界=%.1f (预期100), 南界=%.1f (预期200)", r.West(), r.South())
	}
	if r.East() != 200.0 || r.North() != 250.0 {
		t.Errorf("边界计算错误: 东界=%.1f (预期200), 北界=%.1f (预期250)", r.East(), r.North())
	}

	// 测试坐标转换
	testCases := []struct {
		col, row int
		x, y     float64
	}{
		{0, 0, 102.5, 247.5},  // 左上角 (行0,列0)
		{19, 0, 197.5, 247.5}, // 右上角 (行0,列19)
		{0, 9, 102.5, 202.5},  // 左下角 (行9,列0)
		{19, 9, 197.5, 202.5}, // 右下角 (行9,列19)
		{10, 5, 152.5, 222.5}, // 中心点 (行5,列10) - 修正为正确的中心坐标
	}

	for _, tc := range testCases {
		// 列号转X坐标
		x := r.ColToX(tc.col)
		if math.Abs(x-tc.x) > 1e-5 {
			t.Errorf("ColToX 错误: 列 %d, 预期 %.1f, 实际 %.1f", tc.col, tc.x, x)
		}

		// 行号转Y坐标
		y := r.RowToY(tc.row)
		if math.Abs(y-tc.y) > 1e-5 {
			t.Errorf("RowToY 错误: 行 %d, 预期 %.1f, 实际 %.1f", tc.row, tc.y, y)
		}

		// X坐标转列号
		col := r.XToCol(tc.x)
		if col != tc.col {
			t.Errorf("XToCol 错误: X %.1f, 预期列 %d, 实际列 %d", tc.x, tc.col, col)
		}

		// Y坐标转行号
		row := r.YToRow(tc.y)
		if row != tc.row {
			t.Errorf("YToRow 错误: Y %.1f, 预期行 %d, 实际行 %d", tc.y, tc.row, row)
		}
	}

	// 测试边界外的坐标转换
	if col := r.XToCol(99.9); col != 0 {
		t.Errorf("XToCol 边界外处理错误: 预期 0, 实际 %d", col)
	}
	if col := r.XToCol(200.1); col != 19 {
		t.Errorf("XToCol 边界外处理错误: 预期 19, 实际 %d", col)
	}
	if row := r.YToRow(199.9); row != 9 {
		t.Errorf("YToRow 边界外处理错误: 预期 9, 实际 %d", row)
	}
	if row := r.YToRow(250.1); row != 0 {
		t.Errorf("YToRow 边界外处理错误: 预期 0, 实际 %d", row)
	}
}

// 修复后的栅格坐标转换逻辑
func TestRasterCoordinateCalculation(t *testing.T) {
	r := NewRaster(10, 20, RASTER_DATA_TYPE_FLOAT64)
	r.SetXYPos(100.0, 200.0, 5.0) // 西:100, 南:200, 分辨率:5

	// 验证坐标计算逻辑
	tests := []struct {
		row  int
		y    float64
		desc string
	}{
		{0, 247.5, "最上面一行中心"}, // (10-1-0+0.5)*5 + 200 = 9.5*5+200=247.5
		{1, 242.5, "第二行中心"},   // (10-1-1+0.5)*5 + 200 = 8.5*5+200=242.5
		{5, 222.5, "第六行中心"},   // (10-1-5+0.5)*5 + 200 = 4.5*5+200=222.5
		{9, 202.5, "最下面一行中心"}, // (10-1-9+0.5)*5 + 200 = 0.5*5+200=202.5
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			// 行转Y坐标
			y := r.RowToY(tt.row)
			if math.Abs(y-tt.y) > 1e-5 {
				t.Errorf("RowToY(%d) 错误: 预期 %.1f, 实际 %.1f", tt.row, tt.y, y)
			}

			// Y坐标转行号
			row := r.YToRow(tt.y)
			if row != tt.row {
				t.Errorf("YToRow(%.1f) 错误: 预期行 %d, 实际行 %d", tt.y, tt.row, row)
			}
		})
	}
}

// 测试特定栅格类型 (RasterDouble)
func TestRasterDouble(t *testing.T) {
	// 创建带数据的栅格
	data := make([]float64, 25)
	for i := range data {
		data[i] = float64(i)
	}
	r := NewRasterDoubleWithData(5, 5, data)

	// 验证尺寸
	if r.Rows() != 5 || r.Cols() != 5 {
		t.Errorf("尺寸错误: 预期 5x5, 实际 %dx%d", r.Rows(), r.Cols())
	}

	// 验证值
	for i := 0; i < 25; i++ {
		row := i / 5
		col := i % 5
		val := r.Value(row, col)
		if val != float64(i) {
			t.Errorf("值错误: 位置 (%d,%d) 预期 %.0f, 实际 %.0f", row, col, float64(i), val)
		}
	}

	// 测试填充
	r.Fill(99.9)
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			if r.Value(i, j) != 99.9 {
				t.Errorf("填充错误: 位置 (%d,%d) 预期 99.9, 实际 %.1f", i, j, r.Value(i, j))
			}
		}
	}

	// 测试设置无效值
	r.SetValue(2, 2, math.NaN())
	if !math.IsNaN(r.Value(2, 2)) {
		t.Error("NaN 值处理错误")
	}
}

// 测试特定栅格类型 (RasterChar)
func TestRasterChar(t *testing.T) {
	// 创建带数据的栅格
	data := make([]int8, 9)
	for i := range data {
		data[i] = int8(i)
	}
	r := NewRasterCharWithData(3, 3, data)

	// 验证尺寸
	if r.Rows() != 3 || r.Cols() != 3 {
		t.Errorf("尺寸错误: 预期 3x3, 实际 %dx%d", r.Rows(), r.Cols())
	}

	// 验证值
	for i := 0; i < 9; i++ {
		row := i / 3
		col := i % 3
		val := r.Value(row, col)
		if val != int8(i) {
			t.Errorf("值错误: 位置 (%d,%d) 预期 %d, 实际 %d", row, col, i, val)
		}
	}

	// 测试行读取
	rowData := r.GetRow(1)
	if len(rowData) != 3 {
		t.Errorf("行数据长度错误: 预期 3, 实际 %d", len(rowData))
	}
	if rowData[0] != 3 || rowData[1] != 4 || rowData[2] != 5 {
		t.Errorf("行数据内容错误: 预期 [3,4,5], 实际 [%d,%d,%d]", rowData[0], rowData[1], rowData[2])
	}

	// 测试填充
	r.Fill(1)
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if r.Value(i, j) != 1 {
				t.Errorf("填充错误: 位置 (%d,%d) 预期 1, 实际 %d", i, j, r.Value(i, j))
			}
		}
	}
}

// 测试特定栅格类型 (RasterInt)
func TestRasterInt(t *testing.T) {
	// 创建栅格
	r := NewRasterInt(4, 4, -999)

	// 验证尺寸和无效值
	if r.Rows() != 4 || r.Cols() != 4 {
		t.Errorf("尺寸错误: 预期 4x4, 实际 %dx%d", r.Rows(), r.Cols())
	}
	if r.Value(0, 0) != -999 {
		t.Errorf("无效值初始化错误: 预期 -999, 实际 %d", r.Value(0, 0))
	}

	// 设置值
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			r.SetValue(i, j, int32(i*10+j))
		}
	}

	// 验证值
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			val := r.Value(i, j)
			if val != int32(i*10+j) {
				t.Errorf("值错误: 位置 (%d,%d) 预期 %d, 实际 %d", i, j, i*10+j, val)
			}
		}
	}

	// 测试行读取
	rowData := r.GetRow(2)
	if len(rowData) != 4 {
		t.Errorf("行数据长度错误: 预期 4, 实际 %d", len(rowData))
	}
	if rowData[0] != 20 || rowData[1] != 21 || rowData[2] != 22 || rowData[3] != 23 {
		t.Errorf("行数据内容错误: 预期 [20,21,22,23], 实际 [%d,%d,%d,%d]",
			rowData[0], rowData[1], rowData[2], rowData[3])
	}
}

func TestHemisphericGrid(t *testing.T) {
	r := NewRaster(3, 3, RASTER_DATA_TYPE_FLOAT64)
	r.SetXYPos(0, 0, 10.0)
	r.Hemlines = true

	// 所有列都应该映射到列1
	expectedX := 15.0

	testCases := []int{0, 1, 2}
	for _, col := range testCases {
		x := r.ColToX(col)
		if math.Abs(x-expectedX) > 1e-5 {
			t.Errorf("ColToX 半球网格处理错误: 列 %d, 预期 %.1f, 实际 %.1f", col, expectedX, x)
		}
	}
}

// 测试顶点生成
func TestToVertices(t *testing.T) {
	// 创建栅格
	r := NewRaster(2, 3, RASTER_DATA_TYPE_FLOAT64)
	r.SetXYPos(10.0, 20.0, 5.0)

	// 设置一些值
	r.SetValue(0, 0, 1.0)
	r.SetValue(0, 1, 2.0)
	r.SetValue(0, 2, 3.0)
	r.SetValue(1, 0, 4.0)
	r.SetValue(1, 1, 5.0)
	r.SetValue(1, 2, 6.0)

	// 收集顶点
	var vertices []struct {
		x, y, z float64
	}
	r.ToVertices(func(x, y float64, v interface{}) {
		z := v.(float64)
		vertices = append(vertices, struct{ x, y, z float64 }{x, y, z})
	})

	// 验证顶点 - 更新预期坐标
	expected := []struct {
		x, y, z float64
	}{
		{10.0, 25.0, 1.0}, // 第0行，第0列 (左上角)
		{15.0, 25.0, 2.0}, // 第0行，第1列
		{20.0, 25.0, 3.0}, // 第0行，第2列
		{10.0, 20.0, 4.0}, // 第1行，第0列
		{15.0, 20.0, 5.0}, // 第1行，第1列
		{20.0, 20.0, 6.0}, // 第1行，第2列
	}

	if len(vertices) != len(expected) {
		t.Fatalf("顶点数量错误: 预期 %d, 实际 %d", len(expected), len(vertices))
	}

	for i, v := range vertices {
		if math.Abs(v.x-expected[i].x) > 1e-5 ||
			math.Abs(v.y-expected[i].y) > 1e-5 ||
			math.Abs(v.z-expected[i].z) > 1e-5 {
			t.Errorf("顶点 %d 错误: 预期 (%.1f,%.1f,%.1f), 实际 (%.1f,%.1f,%.1f)",
				i, expected[i].x, expected[i].y, expected[i].z,
				v.x, v.y, v.z)
		}
	}
}

// 测试无效值处理
func TestNoDataHandling(t *testing.T) {
	// 创建带无效值的栅格
	r := NewRasterWithNoData(3, 3, math.NaN())
	r.SetXYPos(0, 0, 1.0)

	// 设置一些有效值
	r.SetValue(1, 1, 10.0)

	// 验证无效值处理
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if i == 1 && j == 1 {
				if IsNaN(r.Value(i, j)) {
					t.Error("有效值被错误标记为无效")
				}
			} else {
				if !IsNaN(r.Value(i, j)) {
					t.Errorf("位置 (%d,%d) 应该是无效值", i, j)
				}
			}
		}
	}

	// 测试坐标转换中的无效值
	col := r.XToCol(1.5)
	row := r.YToRow(1.5)
	if IsNaN(r.Value(row, col)) {
		t.Error("无效值坐标转换处理错误")
	}
}
