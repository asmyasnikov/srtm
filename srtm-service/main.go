package main

import (
	"github.com/asmyasnikov/srtm"
	geojson "github.com/paulmach/go.geojson"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
	"net/http/pprof"
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
}

func main() {
	data, err := srtm.Init(lruCacheSize(), tileDirectory())
	if err != nil {
		log.Error().Caller().Err(err).Msg("")
		return
	}
	defer data.Destroy()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleAddElevations(w, r, data)
	})
	if debug() {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}
	handler := cors.Default().Handler(mux)
	if err := http.ListenAndServe(":"+strconv.Itoa(httpPort()), handler); err != nil {
		log.Error().Caller().Err(err).Msg("")
	}
}

func handleAddElevations(w http.ResponseWriter, r *http.Request, data *srtm.SRTM) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var geoJson = geojson.Geometry{}
	if err := geoJson.UnmarshalJSON(body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := data.AddElevations(&geoJson, true); err != nil {
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
