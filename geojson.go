package srtm

import (
	geojson "github.com/paulmach/go.geojson"
	"github.com/rs/zerolog/log"
	"runtime"
	"sync"
	"time"
)

// AddElevation returns point with 3 coordinates: [longitude, latitude, elevation]
// Param tileDir - directory of hgt-tiles
// Param point - [longitude, latitude]
func (d *SRTM) AddElevation(point []float64) ([]float64, error) {
	ll := LatLng{
		Latitude:  point[1],
		Longitude: point[0],
	}
	tile, err := d.loadTile(ll)
	if err != nil {
		log.Error().Caller().Err(err).Msgf("loadTile: latLng = %s -> error %s", ll.String(), err.Error())
		return nil, err
	}
	tile.setLRU(time.Now())
	elevation, err := tile.GetElevation(ll)
	if err != nil {
		log.Error().Caller().Err(err).Msgf("GetElevation: latLng = %s -> error %s", ll.String(), err.Error())
		return nil, err
	}
	return append(point[:2], float64(elevation)), nil
}

// AddElevations returns geojson with added third coordinate (elevation)
// Param tileDir - directory of hgt-tiles
// Param geoJson - geojson for processing
// Param skipErrors - if false AddElevations use premature exit (on first bad point in geojson). if true all points will be process but bad point will not to be contains elevation coordinate
func (d *SRTM) AddElevations(geoJson *geojson.Geometry, skipErrors bool) error {
	switch geoJson.Type {
	case geojson.GeometryPoint:
		point, err := d.AddElevation(geoJson.Point)
		if err != nil && !skipErrors {
			return err
		}
		geoJson.Point = point
		return nil
	case geojson.GeometryLineString:
		d.process2(geoJson.LineString, runtime.NumCPU())
		return nil
	case geojson.GeometryMultiPoint:
		d.process2(geoJson.MultiPoint, runtime.NumCPU())
		return nil
	case geojson.GeometryPolygon:
		d.process3(geoJson.Polygon, runtime.NumCPU())
		return nil
	case geojson.GeometryMultiLineString:
		d.process3(geoJson.MultiLineString, runtime.NumCPU())
		return nil
	default:
		return nil
	}
}

func (d *SRTM) process3(slice [][][]float64, n int) {
	wg := sync.WaitGroup{}
	wg.Add(n+1)
	type p struct {
		i int
		j int
	}
	ch := make(chan p, n)
	go func() {
		for i := range slice {
			for j := range slice[i] {
				ch <- p{i, j}
			}
		}
		close(ch)
		wg.Done()
	}()
	for i := 0; i < n; i++ {
		go func() {
			for p := range ch {
				point, err := d.AddElevation(slice[p.i][p.j])
				if err != nil {
					log.Error().Caller().Err(err).Msg("")
				}
				slice[p.i][p.j] = point
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func (d *SRTM) process2(slice [][]float64, n int) {
	wg := sync.WaitGroup{}
	wg.Add(n+1)
	ch := make(chan int, n)
	go func() {
		for i := range slice {
			ch <- i
		}
		close(ch)
		wg.Done()
	}()
	for i := 0; i < n; i++ {
		go func() {
			for i := range ch {
				point, err := d.AddElevation(slice[i])
				if err != nil {
					log.Error().Caller().Err(err).Msg("")
				}
				slice[i] = point
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

