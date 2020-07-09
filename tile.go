package srtm

import (
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"github.com/rs/zerolog/log"
	"math"
	"os"
	"path"
	"strings"
	"sync"
)

var cache *lru.Cache
var mtx sync.Mutex
var tileDirectory = "./data"
var storeInMemory = false
var parallel = true

func init() {
	c, err := lru.NewWithEvict(1000, func(key interface{}, value interface{}) {
		log.Debug().Caller().Msgf("remove tile '%s' from cache", key.(string))
	})
	if err != nil {
		panic(err)
	}
	cache = c
}

// Init make initialization of cache
func Init(lruCacheSize int, tileDir string, storeInMemoryMode, parallelMode bool) {
	log.Info().Caller().Int("LRU cache size", lruCacheSize).Bool("store in memory", storeInMemoryMode).Bool("parallel", parallelMode).Msg("")
	_ = cache.Resize(lruCacheSize)
	tileDirectory = tileDir
	storeInMemory = storeInMemoryMode
	parallel = parallelMode
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

func loadTile(ll LatLng) (*Tile, error) {
	key := tileKey(ll)
	mtx.Lock()
	defer mtx.Unlock()
	t, ok := cache.Get(key)
	if ok {
		return t.(*Tile), nil
	}
	tPath, info, err := tilePath(tileDirectory, key, ll)
	if err != nil {
		return nil, err
	}
	if storeInMemory || strings.HasSuffix(tPath, ".gz") {
		sw, size, elevations, err := ReadFile(tPath)
		if err != nil {
			return nil, err
		}
		t = &Tile{
			f:          nil,
			sw:         sw,
			size:       size,
			elevations: elevations,
		}
		if evicted := cache.Add(key, t); evicted {
			log.Error().Caller().Err(err).Msgf("add tile '%s' to cache with evict oldest", key)
		}
		log.Debug().Caller().Str("tile path", tPath).Msg("load tile to memory")
		return t.(*Tile), nil
	}
	sw, size, err := Meta(tPath, info.Size())
	if err != nil {
		return nil, err
	}
	t = &Tile{
		f:          newFileReader(tPath),
		sw:         sw,
		size:       size,
		elevations: nil,
	}
	if evicted := cache.Add(key, t); evicted {
		log.Debug().Caller().Err(err).Msgf("add tile '%s' to cache with evict oldest", key)
	}
	log.Debug().Caller().Str("tile path", tPath).Msg("lazy load tile")
	return t.(*Tile), nil
}

// Tile struct contains hgt-tile meta-data and raw elevations slice
type Tile struct {
	f          *FileReader
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
		log.Error().Caller().Msgf("normalize: error value %d of %s", v, description)
		return 0
	}
	if v > max {
		log.Error().Caller().Msgf("normalize: error value %d of %s", v, description)
		return max
	}
	return v
}

func (t *Tile) quadRowCol(row1, col1, row2, col2, row3, col3, row4, col4 int) (int16, int16, int16, int16) {
	idx1 := (t.size-t.normalize(row1, (t.size-1), "row idx1")-1)*t.size + t.normalize(col1, t.size, "col idx1")
	idx2 := (t.size-t.normalize(row2, (t.size-1), "row idx2")-1)*t.size + t.normalize(col2, t.size, "col idx2")
	idx3 := (t.size-t.normalize(row3, (t.size-1), "row idx3")-1)*t.size + t.normalize(col3, t.size, "col idx3")
	idx4 := (t.size-t.normalize(row4, (t.size-1), "row idx4")-1)*t.size + t.normalize(col4, t.size, "col idx4")
	if t.elevations != nil {
		return t.elevations[idx1], t.elevations[idx2], t.elevations[idx3], t.elevations[idx4]
	}
	err := t.f.open()
	if err != nil {
		log.Error().Caller().Err(err).Msg("")
		return 0, 0, 0, 0
	}
	defer t.f.close()
	e1, err := t.f.elevation(idx1)
	if err != nil {
		log.Error().Caller().Err(err).Int("row", row1).Int("col", col1).Int("idx", idx1).Msg("")
	}
	e2, err := t.f.elevation(idx2)
	if err != nil {
		log.Error().Caller().Err(err).Int("row", row2).Int("col", col2).Int("idx", idx2).Msg("")
	}
	e3, err := t.f.elevation(idx3)
	if err != nil {
		log.Error().Caller().Err(err).Int("row", row3).Int("col", col3).Int("idx", idx3).Msg("")
	}
	e4, err := t.f.elevation(idx4)
	if err != nil {
		log.Error().Caller().Err(err).Int("row", row4).Int("col", col4).Int("idx", idx4).Msg("")
	}
	return e1, e2, e3, e4
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
