package tin

import "math"

type RasterType int

const (
	RASTER_DATA_TYPE_INT8    = 0
	RASTER_DATA_TYPE_UINT8   = 1
	RASTER_DATA_TYPE_INT16   = 2
	RASTER_DATA_TYPE_UINT16  = 3
	RASTER_DATA_TYPE_INT32   = 4
	RASTER_DATA_TYPE_UINT32  = 5
	RASTER_DATA_TYPE_INT64   = 6
	RASTER_DATA_TYPE_UINT64  = 7
	RASTER_DATA_TYPE_FLOAT32 = 8
	RASTER_DATA_TYPE_FLOAT64 = 9
)

type Raster struct {
	Size      [2]int
	Bounds    [4]float64
	NoData    interface{}
	Type      int32
	Data      interface{}
	pos       [2]float64
	cellsize  float64
	transform func(*Vertex) Vertex
}

func NewRasterWithNoData(row, column int, noData interface{}) *Raster {
	r := Raster{}
	r.Size = [2]int{row, column}
	switch t := noData.(type) {
	case int8:
		r.Type = RASTER_DATA_TYPE_INT8
		r.NoData = t
		ds := make([]int8, row*column)
		for i := 0; i < row*column; i++ {
			ds[i] = t
		}
		r.Data = ds
	case uint8:
		r.Type = RASTER_DATA_TYPE_UINT8
		r.NoData = t
		ds := make([]uint8, row*column)
		for i := 0; i < row*column; i++ {
			ds[i] = t
		}
		r.Data = ds
	case int16:
		r.Type = RASTER_DATA_TYPE_INT16
		r.NoData = t
		ds := make([]int16, row*column)
		for i := 0; i < row*column; i++ {
			ds[i] = t
		}
		r.Data = ds
	case uint16:
		r.Type = RASTER_DATA_TYPE_UINT16
		r.NoData = t
		ds := make([]uint16, row*column)
		for i := 0; i < row*column; i++ {
			ds[i] = t
		}
		r.Data = ds
	case int32:
		r.Type = RASTER_DATA_TYPE_INT32
		r.NoData = t
		ds := make([]int32, row*column)
		for i := 0; i < row*column; i++ {
			ds[i] = t
		}
		r.Data = ds
	case uint32:
		r.Type = RASTER_DATA_TYPE_UINT32
		r.NoData = t
		ds := make([]uint32, row*column)
		for i := 0; i < row*column; i++ {
			ds[i] = t
		}
		r.Data = ds
	case int64:
		r.Type = RASTER_DATA_TYPE_INT64
		r.NoData = t
		ds := make([]int64, row*column)
		for i := 0; i < row*column; i++ {
			ds[i] = t
		}
		r.Data = ds
	case uint64:
		r.Type = RASTER_DATA_TYPE_UINT64
		r.NoData = t
		ds := make([]uint64, row*column)
		for i := 0; i < row*column; i++ {
			ds[i] = t
		}
		r.Data = ds
	case float32:
		r.Type = RASTER_DATA_TYPE_FLOAT32
		r.NoData = t
		ds := make([]float32, row*column)
		for i := 0; i < row*column; i++ {
			ds[i] = t
		}
		r.Data = ds
	case float64:
		r.Type = RASTER_DATA_TYPE_FLOAT64
		r.NoData = t
		ds := make([]float64, row*column)
		for i := 0; i < row*column; i++ {
			ds[i] = t
		}
		r.Data = ds
	}
	return &r
}

func NewRaster(row, column, dataType int) *Raster {
	r := Raster{}
	r.Size = [2]int{row, column}
	r.Type = int32(dataType)
	r.NoData = -400.0
	switch dataType {
	case RASTER_DATA_TYPE_INT8:
		r.Data = make([]int8, row*column)
	case RASTER_DATA_TYPE_UINT8:
		r.Data = make([]uint8, row*column)
	case RASTER_DATA_TYPE_INT16:
		r.Data = make([]int16, row*column)
	case RASTER_DATA_TYPE_UINT16:
		r.Data = make([]uint16, row*column)
	case RASTER_DATA_TYPE_INT32:
		r.Data = make([]int32, row*column)
	case RASTER_DATA_TYPE_UINT32:
		r.Data = make([]uint32, row*column)
	case RASTER_DATA_TYPE_INT64:
		r.Data = make([]int64, row*column)
	case RASTER_DATA_TYPE_UINT64:
		r.Data = make([]uint64, row*column)
	case RASTER_DATA_TYPE_FLOAT32:
		r.Data = make([]float32, row*column)
	case RASTER_DATA_TYPE_FLOAT64:
		r.Data = make([]float64, row*column)
	}
	return &r
}

func NewRasterWithData(row, column int, data interface{}) *Raster {
	r := Raster{}
	r.Size = [2]int{row, column}
	r.NoData = -400.0

	switch t := data.(type) {
	case []int8:
		r.Type = RASTER_DATA_TYPE_INT8
		r.Data = t
	case []uint8:
		r.Type = RASTER_DATA_TYPE_UINT8
		r.Data = t
	case []int16:
		r.Type = RASTER_DATA_TYPE_INT16
		r.Data = t
	case []uint16:
		r.Type = RASTER_DATA_TYPE_UINT16
		r.Data = t
	case []int32:
		r.Type = RASTER_DATA_TYPE_INT32
		r.Data = t
	case []uint32:
		r.Type = RASTER_DATA_TYPE_UINT32
		r.Data = t
	case []int64:
		r.Type = RASTER_DATA_TYPE_INT64
		r.Data = t
	case []uint64:
		r.Type = RASTER_DATA_TYPE_UINT64
		r.Data = t
	case []float32:
		r.Type = RASTER_DATA_TYPE_FLOAT32
		r.Data = t
	case []float64:
		r.Type = RASTER_DATA_TYPE_FLOAT64
		r.Data = t
	}
	return &r
}

func (r *Raster) SetXYPos(x, y, res float64) {
	r.setCellSize(res)
	r.setPosX(x)
	r.setPosY(y)
}

func (r *Raster) SetTransform(trans func(*Vertex) Vertex) { r.transform = trans }

func (r *Raster) setPosX(xpos float64) { r.pos[0] = xpos }

func (r *Raster) setPosY(ypos float64) { r.pos[1] = ypos }

func (r *Raster) setCellSize(cs float64) { r.cellsize = cs }

func (r *Raster) Rows() int {
	return r.Size[0]
}

func (r *Raster) Cols() int {
	return r.Size[1]
}

func (r *Raster) Count() int {
	return r.Size[0] * r.Size[1]
}

func (r *Raster) North() float64 {
	return r.Bounds[0]
}

func (r *Raster) South() float64 {
	return r.Bounds[1]
}

func (r *Raster) East() float64 {
	return r.Bounds[2]
}

func (r *Raster) West() float64 {
	return r.Bounds[3]
}

func (r *Raster) Value(row, column int) interface{} {
	switch t := r.Data.(type) {
	case []int8:
		return t[row*r.Cols()+column]
	case []uint8:
		return t[row*r.Cols()+column]
	case []int16:
		return t[row*r.Cols()+column]
	case []uint16:
		return t[row*r.Cols()+column]
	case []int32:
		return t[row*r.Cols()+column]
	case []uint32:
		return t[row*r.Cols()+column]
	case []int64:
		return t[row*r.Cols()+column]
	case []uint64:
		return t[row*r.Cols()+column]
	case []float32:
		return t[row*r.Cols()+column]
	case []float64:
		return t[row*r.Cols()+column]
	}
	return r.NoData
}

func (r *Raster) SetValue(row, column int, data interface{}) {
	switch t := r.Data.(type) {
	case []int8:
		t[row*r.Cols()+column] = data.(int8)
	case []uint8:
		t[row*r.Cols()+column] = data.(uint8)
	case []int16:
		t[row*r.Cols()+column] = data.(int16)
	case []uint16:
		t[row*r.Cols()+column] = data.(uint16)
	case []int32:
		t[row*r.Cols()+column] = data.(int32)
	case []uint32:
		t[row*r.Cols()+column] = data.(uint32)
	case []int64:
		t[row*r.Cols()+column] = data.(int64)
	case []uint64:
		t[row*r.Cols()+column] = data.(uint64)
	case []float32:
		t[row*r.Cols()+column] = data.(float32)
	case []float64:
		t[row*r.Cols()+column] = data.(float64)
	}
}

func (r *Raster) GetRow(row int) interface{} {
	switch t := r.Data.(type) {
	case []int8:
		return t[row*r.Cols() : (row+1)*r.Cols()]
	case []uint8:
		return t[row*r.Cols() : (row+1)*r.Cols()]
	case []int16:
		return t[row*r.Cols() : (row+1)*r.Cols()]
	case []uint16:
		return t[row*r.Cols() : (row+1)*r.Cols()]
	case []int32:
		return t[row*r.Cols() : (row+1)*r.Cols()]
	case []uint32:
		return t[row*r.Cols() : (row+1)*r.Cols()]
	case []int64:
		return t[row*r.Cols() : (row+1)*r.Cols()]
	case []uint64:
		return t[row*r.Cols() : (row+1)*r.Cols()]
	case []float32:
		return t[row*r.Cols() : (row+1)*r.Cols()]
	case []float64:
		return t[row*r.Cols() : (row+1)*r.Cols()]
	}
	return nil
}

func (r *Raster) col2x(c int) float64 { return r.pos[0] + (float64(c)+0.5)*r.cellsize }

func (r *Raster) x2col(x float64) int {
	if r.cellsize > 0 {
		return (int)(0.5 + ((x - r.pos[0] - 0.5*r.cellsize) / r.cellsize))
	}
	return 0
}

func (r *Raster) y2row(y float64) int {
	if r.cellsize > 0 {
		rll := (int)(0.5 + (y-r.pos[1]-0.5*r.cellsize)/r.cellsize)
		rtl := r.Rows() - rll - 1
		return rtl
	}
	return 0
}

func (r *Raster) row2y(rtl int) float64 {
	rll := r.Rows() - 1 - rtl
	return r.pos[1] + (float64(rll)+0.5)*r.cellsize
}

func (r *Raster) rowll2y(rll int) float64 {
	return r.pos[1] + (float64(rll)+0.5)*r.cellsize
}

func (r *Raster) colll2x(c int) float64 { return r.col2x(c) }

type VertexReceiverFn func(x, y float64, v interface{})

func (r *Raster) toVertices(receiverFn VertexReceiverFn) {
	cs := r.cellsize
	xpos := r.pos[0]
	ypos := r.pos[1]
	width := r.Cols()
	height := r.Rows()

	for row := 0; row < height; row++ {
		yCoordinate := ypos + (float64(height)-float64(row)-1)*cs
		for c := 0; c < width; c++ {
			xCoordinate := xpos + float64(c)*cs
			receiverFn(xCoordinate, yCoordinate, r.Value(row, c))
		}
	}
}

type RasterDouble struct {
	Raster
}

func NewRasterDouble(row, column int, noData float64) *RasterDouble {
	r := RasterDouble{}
	r.Size = [2]int{row, column}
	r.Type = int32(RASTER_DATA_TYPE_FLOAT64)
	r.NoData = noData
	dt := make([]float64, row*column)
	for i := range dt {
		dt[i] = noData
	}
	r.Data = dt
	return &r
}

func NewRasterDoubleWithData(row, column int, data []float64) *RasterDouble {
	r := RasterDouble{}
	r.Size = [2]int{row, column}
	r.Type = int32(RASTER_DATA_TYPE_FLOAT64)
	r.Data = data
	r.NoData = math.NaN()
	return &r
}

func (r *RasterDouble) Value(row, column int) float64 {
	switch t := r.Data.(type) {
	case []float64:
		return t[row*r.Cols()+column]
	}
	return r.NoData.(float64)
}

func (r *RasterDouble) SetValue(row, column int, data float64) {
	switch t := r.Data.(type) {
	case []float64:
		t[row*r.Cols()+column] = data
	}
}

func (r *RasterDouble) GetRow(row int) []float64 {
	switch t := r.Data.(type) {
	case []float64:
		return t[row*r.Cols() : (row+1)*r.Cols()]
	}
	return nil
}

type RasterChar struct {
	Raster
}

func NewRasterChar(row, column int, noData int8) *RasterChar {
	r := RasterChar{}
	r.Size = [2]int{row, column}
	r.Type = int32(RASTER_DATA_TYPE_INT8)
	r.NoData = noData
	dt := make([]int8, row*column)
	for i := range dt {
		dt[i] = noData
	}
	r.Data = dt
	return &r
}

func NewRasterCharWithData(row, column int, data []int8) *RasterChar {
	r := RasterChar{}
	r.Size = [2]int{row, column}
	r.Type = int32(RASTER_DATA_TYPE_INT8)
	r.Data = data
	r.NoData = math.NaN()

	return &r
}

func (r *RasterChar) Value(row, column int) int8 {
	switch t := r.Data.(type) {
	case []int8:
		return t[row*r.Cols()+column]
	}
	return r.NoData.(int8)
}

func (r *RasterChar) SetValue(row, column int, data int8) {
	switch t := r.Data.(type) {
	case []int8:
		t[row*r.Cols()+column] = data
	}
}

func (r *RasterChar) GetRow(row int) []int8 {
	switch t := r.Data.(type) {
	case []int8:
		return t[row*r.Cols() : (row+1)*r.Cols()]
	}
	return nil
}

type RasterInt struct {
	Raster
}

func NewRasterInt(row, column int, noData int32) *RasterInt {
	r := RasterInt{}
	r.Size = [2]int{row, column}
	r.Type = int32(RASTER_DATA_TYPE_INT32)
	r.NoData = noData
	dt := make([]int32, row*column)
	for i := range dt {
		dt[i] = noData
	}
	r.Data = dt
	return &r
}

func NewRasterIntWithData(row, column int, data []int32) *RasterInt {
	r := RasterInt{}
	r.Size = [2]int{row, column}
	r.Type = int32(RASTER_DATA_TYPE_INT32)
	r.Data = data
	r.NoData = math.NaN()

	return &r
}

func (r *RasterInt) Value(row, column int) int32 {
	switch t := r.Data.(type) {
	case []int32:
		return t[row*r.Cols()+column]
	}
	return r.NoData.(int32)
}

func (r *RasterInt) SetValue(row, column int, data int32) {
	switch t := r.Data.(type) {
	case []int32:
		t[row*r.Cols()+column] = data
	}
}

func (r *RasterInt) GetRow(row int) []int32 {
	switch t := r.Data.(type) {
	case []int32:
		return t[row*r.Cols() : (row+1)*r.Cols()]
	}
	return nil
}
