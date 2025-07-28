package tin

import (
	"fmt"
	"math"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-geoid"
)

// 新增生成北京附近瓦片边界函数
func GenerateSampleTileBBox(zoom int) vec2d.Rect {
	fwLon := 118.0500 // 经度
	fwLat := 36.8167  // 纬度

	// 转换为Web墨卡托坐标
	srcProj := geo.NewProj(4326) // WGS84 经纬度
	dstProj := geo.NewProj(3857)
	pt, err := transformPoint(srcProj, dstProj, fwLon, fwLat)
	if err != nil {
		panic(err)
	}

	x, y := pt[0], pt[1]

	// 计算瓦片参数
	tileSize := (20037508.34 * 2) / math.Pow(2, float64(zoom))
	tileX := math.Floor(x / tileSize)
	tileY := math.Floor(y / tileSize)

	// 生成边界框
	return vec2d.Rect{
		Min: vec2d.T{
			tileX * tileSize,
			tileY * tileSize,
		},
		Max: vec2d.T{
			(tileX + 1) * tileSize,
			(tileY + 1) * tileSize,
		},
	}
}

// 辅助函数：坐标转换
func transformPoint(srcProj, dstProj geo.Proj, x, y float64) (vec2d.T, error) {
	points := srcProj.TransformTo(dstProj, []vec2d.T{{x, y}})
	if len(points) == 0 {
		return vec2d.T{}, fmt.Errorf("transform point failed")
	}
	return points[0], nil
}

// 创建虚拟DEM
func CreateSampleDEM(bbox vec2d.Rect, spacing float64, maxElev float64) *RasterDouble {
	cols := int(bbox.Width()/spacing) + 1
	rows := int(bbox.Height()/spacing) + 1

	data := make([]float64, cols*rows)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			x := bbox.Min[0] + float64(j)*spacing
			y := bbox.Min[1] + float64(i)*spacing
			// 生成波浪地形
			// 使用分形噪声生成更真实地形
			height := maxElev * (0.5*noise2D(x/500, y/500) +
				0.3*noise2D(x/100, y/100) +
				0.2*noise2D(x/20, y/20))
			data[i*cols+j] = height
		}
	}

	raster := NewRasterDoubleWithData(rows, cols, data)
	raster.SetXYPos(
		bbox.Min[0], // 左下角X坐标
		bbox.Max[1], // 左上角Y坐标
		spacing,
	)

	// 设置坐标转换函数（示例：Web墨卡托转WGS84）
	raster.SetTransform(func(v *Vertex) Vertex {
		// 创建原始坐标系（假设原始是Web墨卡托）
		srcProj := geo.NewProj(3857)
		// 目标坐标系（WGS84经纬度）
		dstProj := geo.NewProj(4326)

		// 执行坐标转换
		pt := srcProj.TransformTo(dstProj, []vec2d.T{{v[0], v[1]}})

		// 高程基准转换（示例：HAE转EGM96）
		g := geoid.NewGeoid(geoid.EGM96, false)
		return Vertex{
			pt[0][0],
			pt[0][1],
			g.ConvertHeight(pt[0][1], pt[0][0], v[2], geoid.GEOIDTOELLIPSOID),
		}
	})

	raster.Bounds = [4]float64{bbox.Min[0], bbox.Max[0], bbox.Min[1], bbox.Max[1]}
	return raster
}

func GenerateSampleMesh(bbox vec2d.Rect, srsCode int) (*Mesh, error) {
	const (
		maxSurfaceElev = 50.0
		spacing        = 10.0 // 采样间距
	)

	// 1. 创建地理参考系
	targetProj := geo.NewProj(srsCode)
	config := &GeoConfig{
		SrcProj: targetProj,
		Datum:   geoid.HAE,
		Offset:  0,
	}

	// 2. 创建虚拟DEM并验证
	dem := CreateSampleDEM(bbox, spacing, maxSurfaceElev)
	if dem == nil || dem.Cols() == 0 || dem.Rows() == 0 {
		return nil, fmt.Errorf("DEM创建失败")
	}

	// 3. 初始化网格生成器
	zemlya := NewZemlyaMesh(config)
	if err := zemlya.LoadRaster(dem); err != nil {
		return nil, fmt.Errorf("栅格加载失败: %v", err)
	}

	// 4. 执行贪婪插入算法
	maxError := 0.5 // 最大允许误差
	zemlya.GreedyInsert(maxError)

	// 5. 转换结果网格并设置地理参考
	resultMesh := zemlya.ToMesh()

	// 6. 验证生成结果
	if resultMesh.Count() == 0 {
		return nil, fmt.Errorf("生成的网格为空")
	}
	if !resultMesh.CheckTin() {
		return nil, fmt.Errorf("生成的网格不符合TIN要求")
	}

	return resultMesh, nil
}

// 辅助哈希函数
func hash(x, y int) int {
	return (x*1619 + y*31337) & 0x7fffffff
}

// 新增二维噪声生成函数
func noise2D(x, y float64) float64 {
	// 网格单元整数部分和小数部分
	xi := int(math.Floor(x))
	yi := int(math.Floor(y))
	xf := x - float64(xi)
	yf := y - float64(yi)

	// 生成四个角点随机梯度向量
	grad := func(hash int) (float64, float64) {
		h := hash & 7
		grad := []float64{1, 1, -1, 1, 1, -1, -1, -1, 1, 0, -1, 0, 0, 1, 0, -1}[h]
		return grad, grad
	}

	// 计算四个角点贡献值
	n0, n1, n2, n3 := 0.0, 0.0, 0.0, 0.0
	t0 := 0.5 - xf*xf - yf*yf
	if t0 >= 0 {
		gx, gy := grad(hash(xi, yi))
		t0 *= t0
		n0 = t0 * t0 * (gx*xf + gy*yf)
	}

	t1 := 0.5 - (xf-1)*(xf-1) - yf*yf
	if t1 >= 0 {
		gx, gy := grad(hash(xi+1, yi))
		t1 *= t1
		n1 = t1 * t1 * (gx*(xf-1) + gy*yf)
	}

	t2 := 0.5 - xf*xf - (yf-1)*(yf-1)
	if t2 >= 0 {
		gx, gy := grad(hash(xi, yi+1))
		t2 *= t2
		n2 = t2 * t2 * (gx*xf + gy*(yf-1))
	}

	t3 := 0.5 - (xf-1)*(xf-1) - (yf-1)*(yf-1)
	if t3 >= 0 {
		gx, gy := grad(hash(xi+1, yi+1))
		t3 *= t3
		n3 = t3 * t3 * (gx*(xf-1) + gy*(yf-1))
	}

	// 归一化到[-1,1]范围
	return 40.0 * (n0 + n1 + n2 + n3)
}

type RasterAdapter struct {
	origin       *RasterDouble
	fullCoverage geo.Coverage
	tileGrid     *geo.TileGrid
	originSR     geo.Proj
}

func NewRasterAdapter(tileGrid *geo.TileGrid, coverage geo.Coverage, origin *RasterDouble) *RasterAdapter {
	return &RasterAdapter{
		origin:       origin,
		tileGrid:     tileGrid,
		fullCoverage: coverage,
		originSR:     tileGrid.Srs, // 克隆以避免外部修改
	}
}

func intersectRect(r *vec2d.Rect, other *vec2d.Rect) *vec2d.Rect {
	// 计算最大最小值
	minX := math.Max(r.Min[0], other.Min[0])
	minY := math.Max(r.Min[1], other.Min[1])
	maxX := math.Min(r.Max[0], other.Max[0])
	maxY := math.Min(r.Max[1], other.Max[1])

	// 检查是否有有效交集
	if minX >= maxX || minY >= maxY {
		return nil
	}

	return &vec2d.Rect{
		Min: vec2d.T{minX, minY},
		Max: vec2d.T{maxX, maxY},
	}
}

func (f *RasterAdapter) generator(bbox vec2d.Rect, zoom int) (*RasterDouble, error) {
	// 验证输入
	if f.origin == nil {
		return nil, fmt.Errorf("origin raster is nil")
	}
	if f.origin.Data == nil {
		return nil, fmt.Errorf("origin raster data is nil")
	}

	origin := f.origin
	cellSize := origin.CellSize()
	if cellSize <= 0 {
		return nil, fmt.Errorf("invalid cell size: %f", cellSize)
	}

	// 计算原始栅格的实际地理范围
	originBBox := vec2d.Rect{
		Min: vec2d.T{f.origin.Bounds[0], f.origin.Bounds[2]}, // 原始最小X,Y
		Max: vec2d.T{f.origin.Bounds[1], f.origin.Bounds[3]}, // 原始最大X,Y
	}
	// 计算请求范围与原始栅格的实际交集
	intersect := intersectRect(&bbox, &originBBox)
	if intersect == nil {
		return nil, fmt.Errorf("bbox %v does not intersect origin raster extent %v", bbox, originBBox)
	}

	// 修正行列号计算（使用浮点数精确计算）
	startCol := int(math.Floor((intersect.Min[0] - originBBox.Min[0]) / cellSize))
	endCol := int(math.Ceil((intersect.Max[0] - originBBox.Min[0]) / cellSize))

	// 修正Y轴方向计算（考虑栅格存储顺序）
	startRow := int(math.Floor((originBBox.Max[1] - intersect.Max[1]) / cellSize))
	endRow := int(math.Ceil((originBBox.Max[1] - intersect.Min[1]) / cellSize))

	// 边界检查与修正
	cols := origin.Cols()
	rows := origin.Rows()

	// 边界保护（确保至少包含一个像元）
	startCol = clamp(startCol, 0, cols-1)
	endCol = clamp(endCol, 1, cols)
	startRow = clamp(startRow, 0, rows-1)
	endRow = clamp(endRow, 1, rows)

	// 计算实际行列数
	subCols := endCol - startCol
	subRows := endRow - startRow
	if subCols <= 0 || subRows <= 0 {
		return nil, fmt.Errorf("无效的子栅格尺寸：%dx%d", subCols, subRows)
	}

	// 计算精确地理边界
	exactMinX := originBBox.Min[0] + float64(startCol)*cellSize
	exactMaxX := exactMinX + float64(subCols)*cellSize
	exactMaxY := originBBox.Max[1] - float64(startRow)*cellSize
	exactMinY := exactMaxY - float64(subRows)*cellSize

	// 一次性复制所有数据（更高效）
	dataSlice := f.origin.DataSlice()
	newData := make([]float64, 0, subRows*subCols)
	for r := startRow; r < endRow; r++ {
		startIdx := r*cols + startCol
		endIdx := startIdx + subCols
		newData = append(newData, dataSlice[startIdx:endIdx]...)
	}

	// 创建新栅格并设置地理参考
	subRaster := NewRasterDoubleWithData(subRows, subCols, newData)

	// 设置精确的地理位置（包含偏移量）
	subRaster.SetXYPos(
		exactMinX, // 左下角X
		exactMaxY, // 左上角Y
		cellSize,
	)
	subRaster.SetTransform(origin.transform)
	subRaster.NoData = origin.NoData
	subRaster.Bounds = [4]float64{exactMinX, exactMaxX, exactMinY, exactMaxY}

	return subRaster, nil
}

// 辅助函数：确保值在[min, max]范围内
func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func (f *RasterAdapter) GetDEM(bbox vec2d.Rect, zoom int) (*RasterDouble, error) {
	// 首先尝试在原始坐标系中处理
	dem, err := f.generator(bbox, zoom)
	if err != nil {
		return nil, fmt.Errorf("DEM generation failed for bbox %v: %w", bbox, err)
	}
	return dem, nil
}

func (f *RasterAdapter) Coverage() (geo.Coverage, error) {
	return f.fullCoverage, nil
}
