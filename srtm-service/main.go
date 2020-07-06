package main

import (
	"fmt"
	srtm "github.com/asmyasnikov/srtm"
	geojson "github.com/paulmach/go.geojson"
	"github.com/rs/cors"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

// TILE_DIRECTORY is a directory of hgt-tiles
var TILE_DIRECTORY = tileDirectory()

// HTTP_PORT - http port of web-service
var HTTP_PORT = httpPort()

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

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleAddElevations)

	handler := cors.Default().Handler(mux)
	if err := http.ListenAndServe(":"+strconv.Itoa(HTTP_PORT), handler); err != nil {
		fmt.Println(err)
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
	geoJson, err = srtm.AddElevations(TILE_DIRECTORY, geoJson, true)
	if err != nil {
		http.Error(w, "can't read body", http.StatusInternalServerError)
		return
	}
	body, err = geoJson.MarshalJSON()
	if err != nil {
		http.Error(w, "can't read body", http.StatusInternalServerError)
		return
	}
	fmt.Println(string(body))
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
