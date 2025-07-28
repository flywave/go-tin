package tin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-geoid"

	vec2d "github.com/flywave/go3d/float64/vec2"
)

// 新增数据获取器接口
type DemProvider interface {
	GetDEM(bbox vec2d.Rect, zoom int) (*RasterDouble, error)
	Coverage() (geo.Coverage, error)
}

type TileExporter interface {
	SaveTile(mesh *Mesh, path string) error
	Extension() string
	RelativeTilePath(zoom, x, y int) string
}

type OBJTileExporter struct{}

func (s *OBJTileExporter) SaveTile(mesh *Mesh, path string) error {
	return mesh.ExportOBJ(path, true)
}

func (s *OBJTileExporter) Extension() string {
	return "obj"
}

func (s *OBJTileExporter) RelativeTilePath(zoom, x, y int) string {
	return filepath.Join(fmt.Sprintf("%d", zoom), fmt.Sprintf("%d", x), fmt.Sprintf("%d.%s", y, s.Extension()))
}

var DefaultTileExporter TileExporter = &OBJTileExporter{}

type TinTilerConfig struct {
	OutputDir   string
	TileGrid    *geo.TileGrid
	MinZoom     int
	MaxZoom     int
	Concurrency int
	MaxError    float64
	Provider    DemProvider // 替换原来的DEMLoader
	Exporter    TileExporter
	Progress    Progress
	Coverage    geo.Coverage
	AutoZoom    bool
	Datum       geoid.VerticalDatum
	Offset      float64
}

type TinTiler struct {
	config        *TinTilerConfig
	taskQueue     chan *tileTask
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	totalTasks    int64
	processed     int64
	firstError    error
	workersCancel context.CancelFunc
	errOnce       sync.Once
	coverage      *vec2d.Rect
	errChan       chan error
}

type tileTask struct {
	zoom int
	x    int
	y    int
}

func NewTinTiler(config *TinTilerConfig) *TinTiler {
	ctx, cancel := context.WithCancel(context.Background())

	if config.Concurrency <= 0 {
		config.Concurrency = 4
	}
	if config.Exporter == nil {
		config.Exporter = DefaultTileExporter
	}
	if config.Progress == nil {
		config.Progress = NewDefaultProgress()
	}
	return &TinTiler{
		config:    config,
		taskQueue: make(chan *tileTask, config.Concurrency*2),
		ctx:       ctx,
		cancel:    cancel,
		errChan:   make(chan error, config.Concurrency),
	}
}

func (t *TinTiler) calculateOptimalZoom(coverage geo.Coverage) error {
	coverageBBox := coverage.GetBBox()
	if !coverage.GetSrs().Eq(t.config.TileGrid.Srs) {
		coverageBBox = coverage.GetSrs().TransformRectTo(
			t.config.TileGrid.Srs,
			coverageBBox,
			16,
		)
	}

	bbox, level, err := t.config.TileGrid.GetAffectedBBoxAndLevel(
		coverageBBox,
		[2]uint32{t.config.TileGrid.TileSize[0], t.config.TileGrid.TileSize[1]},
		t.config.TileGrid.Srs,
	)

	if err == nil {
		t.config.MinZoom = level
		t.coverage = &bbox
		return nil
	}
	return fmt.Errorf("failed to calculate optimal zoom level: %w", err)
}

func (t *TinTiler) preprocess() error {
	if t.config.Coverage == nil {
		coverage, err := t.config.Provider.Coverage()
		if err != nil {
			return err
		}

		t.config.Coverage = coverage
	}

	if t.config.AutoZoom {
		if err := t.calculateOptimalZoom(t.config.Coverage); err != nil {
			return err
		}
	}

	bbox := t.config.Coverage.GetBBox()
	if !t.config.Coverage.GetSrs().Eq(t.config.TileGrid.Srs) {
		bbox = t.config.Coverage.GetSrs().TransformRectTo(
			t.config.TileGrid.Srs,
			bbox,
			16,
		)
	}
	t.coverage = &bbox
	return nil
}

func (t *TinTiler) Run() error {
	defer t.cancel()
	defer close(t.errChan) // 确保错误通道关闭
	defer t.config.Progress.Complete()

	if err := t.preprocess(); err != nil {
		return err
	}

	// 计算总任务量
	t.calculateTotalTasks()
	if t.totalTasks == 0 {
		return fmt.Errorf("no tiles to process in the given bounding box")
	}
	t.config.Progress.Init(int(t.totalTasks)) // 新增初始化方法

	// 创建用于取消工作goroutine的上下文
	workersCtx, workersCancel := context.WithCancel(t.ctx)
	t.workersCancel = workersCancel
	defer workersCancel()

	// 启动工作池
	t.startWorkers(workersCtx)

	// 生成任务
	go t.generateTasks()

	// 等待所有任务完成
	t.wg.Wait()

	// 检查是否有错误发生
	return t.firstError
}

func (t *TinTiler) Stop() {
	t.cancel()
}

func (t *TinTiler) calculateTotalTasks() {
	var total int64
	bbox := t.coverage
	for z := t.config.MinZoom; z <= t.config.MaxZoom; z++ {
		minX, maxX, minY, maxY := t.config.TileGrid.GetAffectedTilesRange(*bbox, z)
		total += int64((maxX - minX + 1) * (maxY - minY + 1))
	}
	atomic.StoreInt64(&t.totalTasks, total)
}

func (t *TinTiler) startWorkers(ctx context.Context) {
	for i := 0; i < t.config.Concurrency; i++ {
		t.wg.Add(1)
		go func() {
			defer t.wg.Done()
			t.worker(ctx)
		}()
	}
}

func (t *TinTiler) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-t.taskQueue:
			if !ok {
				return
			}
			t.processTile(task)
		}
	}
}

func (t *TinTiler) generateTasks() {
	defer close(t.taskQueue)

	for z := t.config.MinZoom; z <= t.config.MaxZoom; z++ {
		bbox := vec2d.Rect{}
		if t.coverage != nil {
			bbox = *t.coverage
		}

		_, _, tileIter, _ := t.config.TileGrid.GetAffectedLevelTiles(bbox, z)

		for {
			x, y, z, done := tileIter.Next()
			if done {
				break
			}

			task := &tileTask{zoom: z, x: x, y: y}

			select {
			case <-t.ctx.Done():
				return
			case t.taskQueue <- task:
			}
		}
	}
}

func (t *TinTiler) processTile(task *tileTask) {
	// 确保即使发生错误也能更新进度
	processed := atomic.AddInt64(&t.processed, 1)
	t.config.Progress.Update(int(processed), int(atomic.LoadInt64(&t.totalTasks)))

	// 获取瓦片地理范围
	tileBBox := t.config.TileGrid.TileBBox([3]int{task.x, task.y, task.zoom}, false)

	// 加载DEM数据
	dem, err := t.config.Provider.GetDEM(tileBBox, task.zoom)
	if err != nil {
		t.reportError(fmt.Errorf("tile %d/%d/%d DEM加载失败: %w", task.zoom, task.x, task.y, err))
		return
	}

	// 生成TIN
	mesh := GenerateTinMesh(dem, t.config.MaxError, &GeoConfig{
		SrcProj: t.config.TileGrid.Srs,
		Datum:   t.config.Datum,
		Offset:  t.config.Offset,
	})

	// 导出瓦片
	relPath := t.config.Exporter.RelativeTilePath(task.zoom, task.x, task.y)
	tilePath := filepath.Join(t.config.OutputDir, relPath)

	if err := os.MkdirAll(filepath.Dir(tilePath), 0755); err != nil {
		t.reportError(fmt.Errorf("创建目录失败: %w", err))
		return
	}

	if err := t.config.Exporter.SaveTile(mesh, tilePath); err != nil {
		t.reportError(fmt.Errorf("保存瓦片失败: %w", err))
	}
}

func (t *TinTiler) reportError(err error) {
	select {
	case t.errChan <- err:
	default:
	}

	select {
	case err := <-t.errChan:
		t.errOnce.Do(func() {
			t.firstError = err
			t.cancel()
		})
	default:
	}
}
