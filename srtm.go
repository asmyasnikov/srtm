package srtm

import (
	"encoding/binary"
	lru "github.com/hashicorp/golang-lru"
	"github.com/rs/zerolog/log"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

// SRTM is a struct contains all internal data
type SRTM struct {
	cache         *lru.Cache
	mtx           sync.Mutex
	tileDirectory string
	done          chan (struct{})
	bads          []string
}

// New make initialization of cache
func New(lruCacheSize int, tileDir string, expiration time.Duration) (*SRTM, error) {
	log.Info().Caller().Int("LRU cache size", lruCacheSize).Str("tile dir", tileDir).Msg("")
	cache, err := lru.NewWithEvict(lruCacheSize, func(key interface{}, value interface{}) {
		log.Debug().Caller().Msgf("remove tile '%s' from cache", key.(string))
		tile, ok := value.(*Tile)
		if !ok {
			log.Error().Caller().Msgf("cache value for key '%s' is not a tile (%+v)", key, value)
			return
		}
		if tile.f != nil {
			if err := tile.f.Close(); err != nil {
				log.Error().Caller().Err(err).Msg("")
			}
		}
		runtime.GC()
	})
	if err != nil {
		log.Error().Caller().Err(err).Msg("")
		return nil, err
	}
	srtm := &SRTM{
		cache:         cache,
		mtx:           sync.Mutex{},
		tileDirectory: tileDir,
		done:          make(chan struct{}),
		bads:          make([]string, 0),
	}
	if expiration > 0 {
		go srtm.sanityCleanLoop(expiration)
	}
	return srtm, nil
}

// Destroy clean all internal data
func (d *SRTM) Destroy() {
	d.mtx.Lock()
	d.cache.Purge()
	close(d.done)
	d.mtx.Unlock()
}

func (d *SRTM) Size() uint64 {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	total := uint64(0)
	for _, key := range d.cache.Keys() {
		value, ok := d.cache.Get(key)
		if !ok {
			continue
		}
		tile, ok := value.(*Tile)
		if !ok {
			continue
		}
		total += uint64(unsafe.Sizeof(*tile))
		total += uint64(unsafe.Sizeof(*tile.sw))
		total += uint64(unsafe.Sizeof(tile.size))
		if tile.f != nil {
			total += uint64(unsafe.Sizeof(*tile.f))
		}
		if len(tile.elevations) > 0 {
			total += uint64(binary.Size(tile.elevations))
		}
	}
	return total
}

func (d *SRTM) sanityCleanLoop(expiration time.Duration) {
	for {
		select {
		case <-time.After(expiration / 2):
			d.sanityClean(expiration)
		case <-d.done:
			return
		}
	}
}

func (d *SRTM) sanityClean(expiration time.Duration) {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	for _, key := range d.cache.Keys() {
		value, ok := d.cache.Get(key)
		if !ok {
			continue
		}
		tile, ok := value.(*Tile)
		if !ok {
			continue
		}
		if time.Since(tile.lru) > expiration {
			d.cache.Remove(key)
		}
	}
}
