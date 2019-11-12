package tin

import "math"

const (
	EarthRadius       = 6378137.0
	MinLatitude       = -85.05112878
	MaxLatitude       = 85.05112878
	MinLongitude      = -180.0
	MaxLongitude      = 180.0
	TileSize          = 256
	MaxLevelOfDetail  = 38
	HalfCircumference = 20037508.342789243076571549020
)

func clip(n, minValue, maxValue float64) float64 {
	if n < minValue {
		return minValue
	}
	if n > maxValue {
		return maxValue
	}
	return n
}

func MapSize(levelOfDetail uint64) uint64 {
	return TileSize << levelOfDetail
}

func LatLongToPixelXY(latitude, longitude float64, levelOfDetail uint64) (pixelX, pixelY int64) {
	latitude = clip(latitude, MinLatitude, MaxLatitude)
	longitude = clip(longitude, MinLongitude, MaxLongitude)
	x := (longitude + 180) / 360
	sinLatitude := math.Sin(latitude * math.Pi / 180)
	y := 0.5 - math.Log((1+sinLatitude)/(1-sinLatitude))/(4*math.Pi)
	mapSize := float64(MapSize(levelOfDetail))
	pixelX = int64(clip(x*mapSize+0.5, 0, mapSize-1))
	pixelY = int64(clip(y*mapSize+0.5, 0, mapSize-1))
	return
}

func PixelXYToLatLong(pixelX, pixelY int64, levelOfDetail uint64) (latitude, longitude float64) {
	mapSize := float64(MapSize(levelOfDetail))
	x := (clip(float64(pixelX), 0, mapSize-1) / mapSize) - 0.5
	y := 0.5 - (clip(float64(pixelY), 0, mapSize-1) / mapSize)
	latitude = 90 - 360*math.Atan(math.Exp(-y*2*math.Pi))/math.Pi
	longitude = 360 * x
	return
}

func PixelXYToTileXY(pixelX, pixelY int64) (tileX, tileY int64) {
	return pixelX >> 8, pixelY >> 8
}

func TileXYToPixelXY(tileX, tileY int64) (pixelX, pixelY int64) {
	return tileX << 8, tileY << 8
}

func TileXYToQuadKey(tileX, tileY int64, levelOfDetail uint64) string {
	quadKey := make([]byte, levelOfDetail)
	for i, j := levelOfDetail, 0; i > 0; i, j = i-1, j+1 {
		mask := int64(1 << (i - 1))
		if (tileX & mask) != 0 {
			if (tileY & mask) != 0 {
				quadKey[j] = '3'
			} else {
				quadKey[j] = '1'
			}
		} else if (tileY & mask) != 0 {
			quadKey[j] = '2'
		} else {
			quadKey[j] = '0'
		}
	}
	return string(quadKey)
}

func QuadKeyToTileXY(quadKey string) (tileX, tileY int64, levelOfDetail uint64) {
	levelOfDetail = uint64(len(quadKey))
	for i := levelOfDetail; i > 0; i-- {
		mask := int64(1 << (i - 1))
		switch quadKey[levelOfDetail-i] {
		case '0':
		case '1':
			tileX |= mask
		case '2':
			tileY |= mask
		case '3':
			tileX |= mask
			tileY |= mask
		default:
			panic("Invalid QuadKey digit sequence.")
		}
	}
	return
}

func PixelXYTToMeters(pixelX, pixelY int64, levelOfDetail uint64) (meterX, meterY float64) {
	invTileSize := 1.0 / TileSize
	dres := 2.0 * HalfCircumference * invTileSize
	res := dres / float64(uint64(1)<<levelOfDetail)
	meterX = float64(pixelX)*res - HalfCircumference
	meterY = float64(pixelY)*res - HalfCircumference
	return meterX, meterY
}

func TileBounds(tileX, tileY int64, levelOfDetail uint64) BBox2d {
	minx, miny := PixelXYTToMeters(tileX*int64(TileSize), tileY*int64(TileSize), levelOfDetail)
	maxx, maxy := PixelXYTToMeters((tileX+1)*int64(TileSize), (tileY+1)*int64(TileSize), levelOfDetail)
	return BBox2d{minx, miny, maxx, maxy}
}
