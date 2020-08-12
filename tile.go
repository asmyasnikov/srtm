package srtm

import (
	"encoding/binary"
	"fmt"
	"github.com/rs/zerolog/log"
	"math"
	"os"
	"path"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

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

func tilePath(tileDir string, ll LatLng) (string, os.FileInfo, error) {
	key := tileKey(ll)
	tilePath := path.Join(tileDir, key)
	for _, s := range suffixes {
		tilePath = tilePath + s
		info, err := os.Stat(tilePath)
		if err == nil || os.IsExist(err) {
			return tilePath, info, nil
		}
	}
	return download(tileDir, ll)
}

func (d *SRTM) loadTile(ll LatLng) (*Tile, error) {
	key := tileKey(ll)
	d.mtx.Lock()
	defer d.mtx.Unlock()
	if sort.SearchStrings(d.bads, key) < len(d.bads) {
		return nil, fmt.Errorf("tile for key '%s' marked as bad", key)
	}
	t, ok := d.cache.Get(key)
	if ok {
		return t.(*Tile), nil
	}
	tPath, info, err := tilePath(d.tileDirectory, ll)
	if err != nil {
		d.bads = append(d.bads, key)
		sort.Strings(d.bads)
		return nil, err
	}
	if strings.HasSuffix(tPath, ".gz") {
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
		if evicted := d.cache.Add(key, t); evicted {
			log.Debug().Caller().Err(err).Msgf("add tile '%s' to cache with evict oldest", key)
		}
		log.Debug().Caller().Str("tile path", tPath).Msg("load tile to memory")
		return t.(*Tile), nil
	}
	sw, size, err := Meta(tPath, info.Size())
	if err != nil {
		return nil, err
	}
	file, err := os.Open(tPath)
	if err != nil {
		return nil, err
	}
	t = &Tile{
		f:          file,
		sw:         sw,
		size:       size,
		elevations: nil,
	}
	if evicted := d.cache.Add(key, t); evicted {
		log.Debug().Caller().Err(err).Msgf("add tile '%s' to cache with evict oldest", key)
	}
	log.Debug().Caller().Str("tile path", tPath).Msg("lazy load tile")
	return t.(*Tile), nil
}

// Tile struct contains hgt-tile meta-data and raw elevations slice
type Tile struct {
	f            *os.File
	sw           *LatLng
	size         int
	elevations   []int16
	internalLRU int64
}

func (t *Tile) setLRU(lru time.Time) {
	atomic.StoreInt64(&t.internalLRU, lru.UnixNano())
}

func (t *Tile) LRU() time.Time{
	u := atomic.LoadInt64(&t.internalLRU)
	return time.Unix(u/1e9, u%1e9)
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

func (t *Tile) elevation(idx int) (int16, error) {
	b := make([]byte, 2)
	n, err := t.f.ReadAt(b, int64(idx)*2)
	if err != nil {
		return 0, err
	}
	if n != 2 {
		return 0, fmt.Errorf("error on read file %s at index %d", t.f.Name(), idx)
	}
	return int16(binary.BigEndian.Uint16(b)), nil
}

func (t *Tile) quadRowCol(row1, col1, row2, col2, row3, col3, row4, col4 int) (int16, int16, int16, int16) {
	idx1 := (t.size-t.normalize(row1, (t.size-1), "row idx1")-1)*t.size + t.normalize(col1, t.size, "col idx1")
	idx2 := (t.size-t.normalize(row2, (t.size-1), "row idx2")-1)*t.size + t.normalize(col2, t.size, "col idx2")
	idx3 := (t.size-t.normalize(row3, (t.size-1), "row idx3")-1)*t.size + t.normalize(col3, t.size, "col idx3")
	idx4 := (t.size-t.normalize(row4, (t.size-1), "row idx4")-1)*t.size + t.normalize(col4, t.size, "col idx4")
	if t.elevations != nil {
		return t.elevations[idx1], t.elevations[idx2], t.elevations[idx3], t.elevations[idx4]
	}
	e1, err := t.elevation(idx1)
	if err != nil {
		log.Error().Caller().Err(err).Int("row", row1).Int("col", col1).Int("idx", idx1).Msg("")
	}
	e2, err := t.elevation(idx2)
	if err != nil {
		log.Error().Caller().Err(err).Int("row", row2).Int("col", col2).Int("idx", idx2).Msg("")
	}
	e3, err := t.elevation(idx3)
	if err != nil {
		log.Error().Caller().Err(err).Int("row", row3).Int("col", col3).Int("idx", idx3).Msg("")
	}
	e4, err := t.elevation(idx4)
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
