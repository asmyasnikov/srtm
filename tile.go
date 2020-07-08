package srtm

import (
	"encoding/binary"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"github.com/rs/zerolog/log"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)

// LRU cache of hgt files
var cache *lru.Cache

var pool = &sync.Pool{
	New: func() interface{} { return make([]byte, 2) },
}

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

func storeInMemoryMode() bool {
	v := os.Getenv("STORE_IN_MEMORY")
	return strings.ToLower(v) != "false"
}

func init() {
	s := lruCacheSize()
	log.Info().Caller().Int("LRU cache size", s).Msg("")
	c, err := lru.NewWithEvict(s, func(key interface{}, value interface{}) {
		log.Debug().Caller().Msgf("remove tile '%s' from cache\n", key.(string))
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
	)
}

var suffixes = []string{
	"",
	".hgt",
	".gz",
}

func tilePath(tileDir string, key string, ll LatLng) (string, os.FileInfo, error) {
	tilePath := path.Join(tileDir, key)
	for _, s := range suffixes {
		tilePath = tilePath + s
		info, err := os.Stat(tilePath)
		if err == nil || os.IsExist(err) {
			return tilePath, info, nil
		}
	}
	return download(tileDir, key, ll)
}

func loadTile(tileDir string, ll LatLng) (*Tile, error) {
	key := tileKey(ll)
	t, ok := cache.Get(key)
	if ok {
		return t.(*Tile), nil
	}
	tPath, info, err := tilePath(tileDir, key, ll)
	if err != nil {
		return nil, err
	}
	if storeInMemoryMode() || strings.HasSuffix(tPath, ".gz") {
		sw, size, elevations, err := ReadFile(tPath)
		if err != nil {
			return nil, err
		}
		t = &Tile{
			file:       tPath,
			sw:         sw,
			size:       size,
			elevations: elevations,
		}
		if evicted := cache.Add(key, t); evicted {
			log.Error().Caller().Err(err).Msgf("add tile '%s' to cache with evict oldest\n", key)
		}
		return t.(*Tile), nil
	}
	sw, size, err := Meta(tPath, info.Size())
	if err != nil {
		return nil, err
	}
	t = &Tile{
		file:       tPath,
		sw:         sw,
		size:       size,
		elevations: nil,
	}
	if evicted := cache.Add(key, t); evicted {
		log.Error().Caller().Err(err).Msgf("add tile '%s' to cache with evict oldest\n", key)
	}
	return t.(*Tile), nil
}

// Tile struct contains hgt-tile meta-data and raw elevations slice
type Tile struct {
	file       string
	sw         *LatLng
	size       int
	elevations []int16
}

// GetElevation returns elevation for lat/lng
func (t *Tile) GetElevation(ll LatLng) (float64, error) {
	size := float64(t.size - 1)
	row := (ll.Latitude - t.sw.Latitude) * size
	col := (ll.Longitude - t.sw.Longitude) * size
	if row < 0 || col < 0 || row > size || col > size {
		return 0, fmt.Errorf("lat/lng is outside tile bounds (row=%f, col=%f, size=%f)", row, col, size)
	}
	return t.interpolate(row, col), nil
}

func avg(v1, v2, f float64) float64 {
	return v1 + (v2-v1)*f
}

func (t *Tile) normalize(v, max int, description string) int {
	if v < 0 {
		log.Error().Caller().Msgf("normalize: error value %d of %s\n", v, description)
		return 0
	}
	if v > max {
		log.Error().Caller().Msgf("normalize: error value %d of %s\n", v, description)
		return max
	}
	return v
}

func (t *Tile) rowCol(row, col int, description string) int16 {
	idx := (t.size-t.normalize(row, (t.size-1), "row "+description)-1)*t.size + t.normalize(col, t.size, "col "+description)
	if t.elevations != nil {
		return t.elevations[idx]
	}
	f, err := os.Open(t.file)
	if err != nil {
		log.Error().Caller().Err(err).Msgf("error on open file %s\n", t.file)
		return 0
	}
	defer f.Close()
	b := pool.Get().([]byte)
	defer pool.Put(b)
	n, err := f.ReadAt(b, int64(idx)*2)
	if n != 2 {
		log.Error().Caller().Err(err).Msgf("error on read file %s at index %d\n", t.file, int64(idx)*2)
		return 0
	}
	return int16(binary.BigEndian.Uint16(b))
}

func (t *Tile) elevation(f *os.File, idx int) int16 {
	b := pool.Get().([]byte)
	defer pool.Put(b)
	n, err := f.ReadAt(b, int64(idx)*2)
	if err != nil {
		log.Error().Caller().Err(err).Msgf("error '%s' on read file %s at index %d\n", err.Error(), t.file, idx)
		return 0
	}
	if n != 2 {
		log.Error().Caller().Err(err).Msgf("error on read file %s at index %d\n", t.file, idx)
		return 0
	}
	return int16(binary.BigEndian.Uint16(b))
}

func (t *Tile) quadRowCol(row1, col1, row2, col2, row3, col3, row4, col4 int) (int16, int16, int16, int16) {
	idx1 := (t.size-t.normalize(row1, (t.size-1), "row idx1")-1)*t.size + t.normalize(col1, t.size, "col idx1")
	idx2 := (t.size-t.normalize(row2, (t.size-1), "row idx2")-1)*t.size + t.normalize(col2, t.size, "col idx2")
	idx3 := (t.size-t.normalize(row3, (t.size-1), "row idx3")-1)*t.size + t.normalize(col3, t.size, "col idx3")
	idx4 := (t.size-t.normalize(row4, (t.size-1), "row idx4")-1)*t.size + t.normalize(col4, t.size, "col idx4")
	if t.elevations != nil {
		return t.elevations[idx1], t.elevations[idx2], t.elevations[idx3], t.elevations[idx4]
	}
	f, err := os.Open(t.file)
	if err != nil {
		log.Error().Caller().Err(err).Msgf("error on open file %s\n", t.file)
		return 0, 0, 0, 0
	}
	defer f.Close()
	return t.elevation(f, idx1), t.elevation(f, idx2), t.elevation(f, idx3), t.elevation(f, idx4)
}

func (t *Tile) interpolate(row, col float64) float64 {
	rowLow := int(math.Floor(row))
	rowHi := rowLow + 1
	rowFrac := row - float64(rowLow)
	colLow := int(math.Floor(col))
	colHi := colLow + 1
	colFrac := col - float64(colLow)
	v00, v10, v11, v01 := t.quadRowCol(rowLow, colLow, rowLow, colHi, rowHi, colHi, rowHi, colLow)
	v1 := avg(float64(v00), float64(v10), colFrac)
	v2 := avg(float64(v01), float64(v11), colFrac)
	return avg(v1, v2, rowFrac)
}
