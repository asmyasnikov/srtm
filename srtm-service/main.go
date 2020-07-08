package main

import (
	srtm "github.com/asmyasnikov/srtm"
	geojson "github.com/paulmach/go.geojson"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// TILE_DIRECTORY is a directory of hgt-tiles
var TILE_DIRECTORY = tileDirectory()

// HTTP_PORT - http port of web-service
var HTTP_PORT = httpPort()

// STORE_IN_MEMORY - store elevation data in memory (all hgt file)
var STORE_IN_MEMORY = storeInMemoryMode()

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
	srtm.Init(lruCacheSize())
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleAddElevations)

	handler := cors.Default().Handler(mux)
	if err := http.ListenAndServe(":"+strconv.Itoa(HTTP_PORT), handler); err != nil {
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
	geoJson, err = srtm.AddElevations(TILE_DIRECTORY, STORE_IN_MEMORY, geoJson, true)
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
