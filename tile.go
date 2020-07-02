package srtm

import (
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"math"
	"os"
	"path"
)

// LRU cache of hgt files
var cache *lru.Cache

func init() {
	c, err := lru.NewWithEvict(1000, func(key interface{}, value interface{}) {
		// if cast value to *Tile failed - panic - it's ok
		value.(*Tile).destroy()
	})
	if err != nil {
		panic(err)
	}
	cache = c
}

func tileKey(ll LatLng) string {
	return fmt.Sprintf("%s%02d%s%03d",
		func() string {
			if ll.Latitude < 0 {
				return "S"
			}
			return "N"
		}(),
		int(math.Abs(math.Floor(ll.Latitude))),
		func() string {
			if ll.Longitude < 0 {
				return "W"
			}
			return "E"
		}(),
		int(math.Abs(math.Floor(ll.Longitude))),
	);
}

var suffixes = []string {
	"",
	".hgt",
	".gz",
}

func tilePath(tileDir string, key string) (string, error) {
	tilePath := path.Join(tileDir, key)
	for _, s := range suffixes {
		tilePath = tilePath + s
		_, err := os.Stat(tilePath)
		if err == nil || os.IsExist(err) {
			return tilePath, nil
		}
	}
	return "", fmt.Errorf("tile file for key = %s is not exists", key)
}

func loadTile(tileDir string, ll LatLng) (*Tile, error) {
	key := tileKey(ll)
	t, ok := cache.Get(key)
	if ok {
		return t.(*Tile), nil
	}
	tPath, err := tilePath(tileDir, key)
	if err != nil {
		return nil, err
	}
	sw, size, elevations, err := ReadFile(tPath)
	if err != nil {
		return nil, err
	}
	t = &Tile{
		sw: sw,
		size: size,
		elevations: elevations,
	}
	_ = cache.Add(key, t)
	return t.(*Tile), nil
}

type Tile struct {
	sw *LatLng
	size int
	elevations []int16
}

func (t *Tile) destroy() {
}

func (t *Tile) getElevation(ll LatLng) (float64, error) {
	row := (ll.Latitude - t.sw.Latitude) * float64(t.size)
	col := (ll.Longitude - t.sw.Longitude) * float64(t.size)
	if row < 0 || col < 0 || row > float64(t.size) || col > float64(t.size) {
		return 0, fmt.Errorf("lat/lng is outside tile bounds (row=%f, col=%f, size=%d)", row, col, t.size)
	}
	return t.interpolate(row, col), nil
}

func avg (v1, v2, f float64) float64 {
	return v1 + (v2 - v1) * f
}

func (t *Tile) rowCol(row, col int) float64 {
	fmt.Println("rowCol:", t.size * row + col, row, col, t.size)
	return float64(t.elevations[t.size * row + col])
}

func (t *Tile) interpolate(row, col float64) float64 {
	rowLow := int(math.Floor(row))
	rowHi := rowLow + 1
	rowFrac := row - float64(rowLow)
	colLow := int(math.Floor(col))
	colHi := colLow + 1
	colFrac := col - float64(colLow)
	v00 := t.rowCol(rowLow, colLow)
	v10 := t.rowCol(rowLow, colHi)
	v11 := t.rowCol(rowHi, colHi)
	v01 := t.rowCol(rowHi, colLow)
	v1 := avg(v00, v10, colFrac)
	v2 := avg(v01, v11, colFrac)
	return avg(v1, v2, rowFrac)
}

