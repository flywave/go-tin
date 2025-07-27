package tin

import (
	"fmt"
	"testing"
)

func TestRaster(t *testing.T) {
	rst := NewRasterDouble(256, 256, 0)
	rst.SetXYPos(0, 0, 1)
	x1 := rst.ColToX(0)
	x2 := rst.ColToX(256)

	y1 := rst.RowToY(0)
	y2 := rst.RowToY(256)
	fmt.Println(x1, "  ", x2, "  ", y1, "  ", y2)
}
