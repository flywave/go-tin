package tin

import (
	"reflect"
	"testing"
)

func TestGenTileScaling(t *testing.T) {
	// 创建测试网格 (2x2平面)
	mesh := &Mesh{
		Vertices: []Vertex{
			{0, 0, 5},
			{10, 0, 5},
			{10, 10, 5},
			{0, 10, 5},
		},
		Faces: []Face{
			{0, 1, 2},
			{0, 2, 3},
		},
		BBox: [2][3]float64{{0, 0, 5}, {10, 10, 5}},
	}

	tm := NewTileMaker(mesh)

	// 测试缩放
	t.Run("WithScaling", func(t *testing.T) {
		scaledMesh, err := tm.GenTile(true)
		if err != nil {
			t.Fatalf("GenTile failed: %v", err)
		}

		// 检查顶点是否归一化到[0,1]
		for _, v := range scaledMesh.Vertices {
			if v[0] < 0 || v[0] > 1 || v[1] < 0 || v[1] > 1 || v[2] != 0 {
				t.Errorf("Vertex %v not normalized correctly", v)
			}
		}

		// 验证边界框
		expectedBBox := [2][3]float64{{0, 0, 0}, {1, 1, 0}}
		if !reflect.DeepEqual(scaledMesh.BBox, expectedBBox) {
			t.Errorf("BBox mismatch. Expected %v, got %v", expectedBBox, scaledMesh.BBox)
		}
	})

	// 测试不缩放
	t.Run("WithoutScaling", func(t *testing.T) {
		unscaledMesh, _ := tm.GenTile(false)
		if !reflect.DeepEqual(unscaledMesh.Vertices, mesh.Vertices) {
			t.Error("Vertices modified when scaling was disabled")
		}
	})
}
