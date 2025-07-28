package tin

import (
	"testing"

	"github.com/flywave/go-geo"
	"github.com/stretchr/testify/assert"
)

// 新增测试用例：验证完整瓦片生成流程
func TestTINTiler(t *testing.T) {
	testBBox := GenerateSampleTileBBox(16)
	dem := CreateSampleDEM(testBBox, 10.0, 100.0)
	coverage := geo.NewBBoxCoverage(testBBox, geo.NewProj(3857), true)
	tileGrid := geo.NewMercTileGrid()

	// 创建Tiler配置
	config := &TinTilerConfig{
		OutputDir:   "tests/tiles",
		TileGrid:    tileGrid,
		MinZoom:     15,
		MaxZoom:     18,
		Concurrency: 2,
		MaxError:    100.0,
		AutoZoom:    true,
		Provider:    NewRasterAdapter(tileGrid, coverage, dem),
		Exporter:    &OBJTileExporter{},
		Coverage:    coverage,
	}

	// 创建并运行瓦片生成器
	tiler := NewTinTiler(config)
	err := tiler.Run()
	assert.NoError(t, err)
}
