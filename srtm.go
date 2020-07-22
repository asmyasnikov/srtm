package srtm

import (
	lru "github.com/hashicorp/golang-lru"
	"github.com/rs/zerolog/log"
	"sync"
)

// SRTM is a struct contains all internal data
type SRTM struct {
	cache         *lru.Cache
	mtx           sync.Mutex
	tileDirectory string
}

// Init make initialization of cache
func Init(lruCacheSize int, tileDir string) (*SRTM, error) {
	c, err := lru.NewWithEvict(lruCacheSize, func(key interface{}, value interface{}) {
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
	})
	if err != nil {
		log.Error().Caller().Err(err).Msg("")
		return nil, err
	}
	log.Info().Caller().Int("LRU cache size", lruCacheSize).Str("tile dir", tileDir).Msg("")
	return &SRTM{
		cache:         c,
		mtx:           sync.Mutex{},
		tileDirectory: tileDir,
	}, nil
}

// Destroy clean all internal data
func (d *SRTM) Destroy() {
	d.mtx.Lock()
	d.cache.Purge()
	d.mtx.Unlock()
}
