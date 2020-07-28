package main

import (
	"flag"
	"fmt"
	"github.com/asmyasnikov/srtm"
	"github.com/dustin/go-humanize"
	"github.com/gorilla/mux"
	geojson "github.com/paulmach/go.geojson"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
	"net/http/pprof"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	flags = map[string]interface{}{
		"debug":          flag.Bool("debug", false, "boolean flag for debug handlers with pprof"),
		"lru-cache-size": flag.Int("lru-cache-size", 1000, "LRU cache size"),
		"www":            flag.String("www", "/", "prefix of handlers"),
		"http-port":      flag.Int("http-port", 80, "http port of web-service"),
		"tile-directory": flag.String("tile-directory", "./data/", "directory of hgt tiles"),
		"log-level":      flag.String("log-level", "error", "logging level"),
		"expiration":     flag.Duration("expiration", time.Minute, "expiration time for tiles in LRU cache"),
	}
	args = map[string]func() interface{}{
		"debug":          debug,
		"lru-cache-size": lruCacheSize,
		"www":            www,
		"http-port":      httpPort,
		"tile-directory": tileDirectory,
		"log-level":      logLevel,
		"expiration":     expiration,
	}
)

func tileDirectory() interface{} {
	v := os.Getenv("TILE_DIRECTORY")
	if len(v) > 0 {
		return v
	}
	tileDirectory := flags["tile-directory"].(*string)
	if tileDirectory != nil {
		return *tileDirectory
	}
	return "./data/"
}

func expiration() interface{} {
	v := os.Getenv("EXPIRATION")
	if len(v) > 0 {
		expiration, err := time.ParseDuration(v)
		if err == nil {
			return expiration
		}
	}
	expiration := flags["expiration"].(*time.Duration)
	if expiration != nil {
		return *expiration
	}
	return time.Minute
}

func httpPort() interface{} {
	v := os.Getenv("HTTP_PORT")
	if len(v) > 0 {
		p, err := strconv.Atoi(v)
		if err == nil {
			return p
		}
	}
	httpPort := flags["http-port"].(*int)
	if httpPort != nil {
		return *httpPort
	}
	return 80
}

func www() interface{} {
	v := os.Getenv("WWW")
	if len(v) > 0 {
		return v
	}
	www := flags["www"].(*string)
	if www != nil {
		return *www
	}
	return "/"
}

func logLevel() interface{} {
	v := os.Getenv("LOG_LEVEL")
	if len(v) > 0 {
		return v
	}
	logLevel := flags["log-level"].(*string)
	if logLevel != nil {
		return *logLevel
	}
	return "error"
}

func lruCacheSize() interface{} {
	v := os.Getenv("LRU_CACHE_SIZE")
	if len(v) > 0 {
		s, err := strconv.Atoi(v)
		if err == nil {
			return s
		}
	}
	lruCacheSize := flags["lru-cache-size"].(*int)
	if lruCacheSize != nil {
		return *lruCacheSize
	}
	return 1000
}

func debug() interface{} {
	v := os.Getenv("DEBUG")
	if len(v) > 0 {
		return strings.ToLower(v) == "true"
	}
	debug := flags["debug"].(*bool)
	if debug != nil {
		return *debug
	}
	return false
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	fmt.Fprintf(os.Stderr, "\nRunning %s with args:\n", os.Args[0])
	for k, v := range args {
		fmt.Fprintf(os.Stderr, "  --%s=%+v\n", k, v())
	}
	fmt.Fprintln(os.Stderr)
}

func main() {
	l, err := zerolog.ParseLevel(logLevel().(string))
	if err != nil {
		log.Error().Caller().Err(err).Msg("")
		return
	}
	zerolog.SetGlobalLevel(l)
	data, err := srtm.New(lruCacheSize().(int), tileDirectory().(string), expiration().(time.Duration))
	if err != nil {
		log.Error().Caller().Err(err).Msg("")
		return
	}
	defer data.Destroy()
	pool := &sync.Pool{
		New: func() interface{} {
			return &geojson.Geometry{}
		},
	}
	router := mux.NewRouter().PathPrefix(www().(string)).Subrouter()
	if debug().(bool) {
		router.HandleFunc("/debug/pprof/", pprof.Index)
		router.HandleFunc("/debug/pprof/allocs", pprof.Handler("allocs").ServeHTTP)
		router.HandleFunc("/debug/pprof/heap", pprof.Handler("heap").ServeHTTP)
		router.HandleFunc("/debug/pprof/block", pprof.Handler("block").ServeHTTP)
		router.HandleFunc("/debug/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP)
		router.HandleFunc("/debug/pprof/mutex", pprof.Handler("mutex").ServeHTTP)
		router.HandleFunc("/debug/pprof/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
		router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		router.HandleFunc("/debug/pprof/profile", pprof.Profile)
		router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		router.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleAddElevations(w, r, data, pool)
	}).Methods(http.MethodPost)
	if debug().(bool) {
		go func() {
			var memory runtime.MemStats
			for {
				runtime.ReadMemStats(&memory)
				log.
					Debug().
					Caller().
					Str("cache size", humanize.Bytes(data.Size())).
					Str("alloc", humanize.Bytes(memory.Alloc)).
					Str("total alloc", humanize.Bytes(memory.TotalAlloc)).
					Str("sys", humanize.Bytes(memory.Sys)).
					Uint32("num gc", memory.NumGC).
					Msg("")
				time.Sleep(time.Second)
			}
		}()
	}
	log.Info().Caller().Int("http-port", httpPort().(int)).Msg("running web-server...")
	if err := http.ListenAndServe(":"+strconv.Itoa(httpPort().(int)), cors.Default().Handler(router)); err != nil {
		log.Error().Caller().Err(err).Msg("")
	}
}

func handleAddElevations(w http.ResponseWriter, r *http.Request, data *srtm.SRTM, pool *sync.Pool) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	geoJson := pool.Get().(*geojson.Geometry)
	defer pool.Put(geoJson)
	if err := geoJson.UnmarshalJSON(body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := data.AddElevations(geoJson, true); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	body, err = geoJson.MarshalJSON()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
