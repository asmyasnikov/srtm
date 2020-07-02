package srtm

import (
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"math"
	"os"
	"path"
	"strconv"
)

// LRU cache of hgt files
var cache *lru.Cache

func lruCacheSize() int {
	v := os.Getenv("LRU_CACHE_SIZE")
	if len(v) == 0 {
		return 1000
	}
	s, err := strconv.Atoi(v)
	if err != nil {
		return 1000
	}
	return s
}


func init() {
	s := lruCacheSize()
	fmt.Println("LRU cache size", s)
	c, err := lru.NewWithEvict(s, func(key interface{}, value interface{}) {
		fmt.Printf("remove tile '%s' from cache\n", key.(string))
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
	if evicted := cache.Add(key, t); evicted {
		fmt.Printf("add tile '%s' to cache with evict oldest\n", key)
	}
	return t.(*Tile), nil
}

type Tile struct {
	sw *LatLng
	size int
	elevations []int16
}

func (t *Tile) getElevation(ll LatLng) (float64, error) {
	size := float64(t.size - 1)
	row := (ll.Latitude - t.sw.Latitude) * size
	col := (ll.Longitude - t.sw.Longitude) * size
	if row < 0 || col < 0 || row > size || col > size {
		return 0, fmt.Errorf("lat/lng is outside tile bounds (row=%f, col=%f, size=%f)", row, col, size)
	}
	return t.interpolate(row, col), nil
}

func avg (v1, v2, f float64) float64 {
	return v1 + (v2 - v1) * f
}

func (t *Tile) normalize(v, max int, description string) int {
	if v < 0 {
		fmt.Printf("normalize: error value %d of %s\n", v, description)
		return 0
	}
	if v > max {
		fmt.Printf("normalize: error value %d of %s\n", v, description)
		return max
	}
	return v
}

func (t *Tile) rowCol(row, col int, description string) float64 {
	return float64(t.elevations[(t.size - t.normalize(row, (t.size-1), "row " + description) - 1) * t.size + t.normalize(col, t.size, "col " + description)])
}

func (t *Tile) interpolate(row, col float64) float64 {
	rowLow := int(math.Floor(row))
	rowHi := int(math.Ceil(row))
	rowFrac := row - float64(rowLow)
	colLow := int(math.Floor(col))
	colHi := int(math.Ceil(col))
	colFrac := col - float64(colLow)
	v00 := t.rowCol(rowLow, colLow, "v00")
	v10 := t.rowCol(rowLow, colHi, "v10")
	v11 := t.rowCol(rowHi, colHi, "v11")
	v01 := t.rowCol(rowHi, colLow, "v01")
	v1 := avg(v00, v10, colFrac)
	v2 := avg(v01, v11, colFrac)
	return avg(v1, v2, rowFrac)
}

