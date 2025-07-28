package tin

import "testing"

// 测试函数
func TestMesh(t *testing.T) {
	// 1. 创建基础三角形
	tri1 := Triangle{
		{0, 0, 0},
		{1, 0, 0},
		{0, 1, 1},
	}
	tri2 := Triangle{
		{1, 0, 0},
		{1, 1, 1},
		{0, 1, 1},
	}

	// 2. 测试初始化
	t.Run("InitFromTriangles", func(t *testing.T) {
		m := Mesh{}
		m.InitFromTriangles([]Triangle{tri1, tri2})

		if len(m.Triangles) != 2 {
			t.Errorf("Expected 2 triangles, got %d", len(m.Triangles))
		}
		if len(m.Vertices) != 0 {
			t.Error("Vertices should be empty after triangle init")
		}
	})

	// 3. 测试添加三角形
	t.Run("AddTriangle", func(t *testing.T) {
		m := Mesh{}
		m.InitFromTriangles(nil)

		m.AddTriangle(tri1, true)
		m.AddTriangle(tri2, true)

		if len(m.Triangles) != 2 {
			t.Errorf("Expected 2 triangles, got %d", len(m.Triangles))
		}
		if len(m.Vertices) != 4 {
			t.Errorf("Expected 4 vertices, got %d", len(m.Vertices))
		}
	})

	// 4. 测试数据转换
	t.Run("DataConversion", func(t *testing.T) {
		m := Mesh{}
		m.InitFromTriangles([]Triangle{tri1, tri2})

		// 转换为分解数据
		m.GenerateDecomposed()
		if len(m.Faces) != 2 || len(m.Vertices) != 4 {
			t.Errorf("Decomposed data invalid: %d faces, %d vertices",
				len(m.Faces), len(m.Vertices))
		}

		// 转换回三角形
		m.GenerateTriangles()
		if len(m.Triangles) != 2 {
			t.Error("Failed to convert back to triangles")
		}
	})

	// 5. 测试边界框
	t.Run("BoundingBox", func(t *testing.T) {
		m := Mesh{}
		m.InitFromTriangles([]Triangle{tri1, tri2})
		m.GenerateDecomposed()

		// 检查自动计算的BBox
		expectedMin := Vertex{0, 0, 0}
		expectedMax := Vertex{1, 1, 1}

		if m.BBox[0] != expectedMin || m.BBox[1] != expectedMax {
			t.Errorf("BBox mismatch.\nExpected: %v-%v\nGot: %v-%v",
				expectedMin, expectedMax, m.BBox[0], m.BBox[1])
		}

		// 测试GetBbox方法
		bbox := m.GetBbox()
		if bbox.Min() != expectedMin || bbox.Max() != expectedMax {
			t.Error("GetBbox returned incorrect values")
		}
	})

	// 6. 测试TIN检查
	t.Run("CheckTin", func(t *testing.T) {
		m := Mesh{}

		// 创建有效TIN
		validTris := []Triangle{
			{{0, 0, 0}, {1, 0, 0}, {0, 1, 0}},
			{{1, 0, 0}, {1, 1, 0}, {0, 1, 0}},
		}
		m.InitFromTriangles(validTris)
		m.GenerateDecomposed()

		if !m.CheckTin() {
			t.Error("Valid TIN failed check")
		}

		// 创建无效TIN（重复顶点）
		invalidTris := []Triangle{
			{{0, 0, 0}, {0, 0, 0}, {1, 1, 0}}, // 退化三角形
		}
		m.InitFromTriangles(invalidTris)
		m.GenerateDecomposed()

		if m.CheckTin() {
			t.Error("Invalid TIN passed check")
		}
	})

	// 7. 测试边缘情况
	t.Run("EdgeCases", func(t *testing.T) {
		m := Mesh{}

		// 空网格测试
		if !m.Empty() || m.Count() != 0 {
			t.Error("Empty mesh not reported as empty")
		}

		// 添加单个三角形
		m.AddTriangle(tri1, false)
		if m.Count() != 1 {
			t.Error("Triangle count incorrect")
		}

		// 清除测试
		m.ClearTriangles()
		if len(m.Triangles) != 0 {
			t.Error("ClearTriangles failed")
		}
	})
}
