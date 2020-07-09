package main

import (
	srtm "github.com/asmyasnikov/srtm"
	geojson "github.com/paulmach/go.geojson"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
)

func tileDirectory() string {
	v := os.Getenv("TILE_DIRECTORY")
	if len(v) == 0 {
		return "./data/"
	}
	return v
}

func httpPort() int {
	v := os.Getenv("HTTP_PORT")
	if len(v) == 0 {
		return 80
	}
	p, err := strconv.Atoi(v)
	if err != nil {
		return 80
	}
	return p
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

func parallel() bool {
	v := os.Getenv("PARALLEL")
	return strings.ToLower(v) != "false"
}

func debug() bool {
	v := os.Getenv("DEBUG")
	return strings.ToLower(v) == "true"
}

func init() {
	logLevel := os.Getenv("LOG_LEVEL")
	if len(logLevel) == 0 {
		logLevel = "error"
	}
	l, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		log.Error().Caller().Err(err).Msg("")
		return
	}
	zerolog.SetGlobalLevel(l)
	srtm.Init(lruCacheSize(), tileDirectory(), storeInMemoryMode(), parallel())
}

func main() {
	if debug() {
		go func() {
			http.ListenAndServe(":6060", nil)
		}()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleAddElevations)

	handler := cors.Default().Handler(mux)
	if err := http.ListenAndServe(":"+strconv.Itoa(httpPort()), handler); err != nil {
		log.Error().Caller().Err(err).Msg("")
	}
}

func handleAddElevations(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	geoJson, err := geojson.UnmarshalGeometry(body)
	if err != nil {
		http.Error(w, "can't unmarshall body", http.StatusBadRequest)
		return
	}
	geoJson, err = srtm.AddElevations(geoJson, true)
	if err != nil {
		http.Error(w, "can't read body", http.StatusInternalServerError)
		return
	}
	body, err = geoJson.MarshalJSON()
	if err != nil {
		http.Error(w, "can't read body", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
