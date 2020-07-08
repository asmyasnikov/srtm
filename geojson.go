package srtm

import (
	geojson "github.com/paulmach/go.geojson"
	"github.com/rs/zerolog/log"
	"sync"
)

// AddElevation returns point with 3 coordinates: [longitude, latitude, elevation]
// Param tileDir - directory of hgt-tiles
// Param point - [longitude, latitude]
func AddElevation(tileDir string, point []float64) ([]float64, error) {
	ll := LatLng{
		Latitude:  point[1],
		Longitude: point[0],
	}
	tile, err := loadTile(tileDir, ll)
	if err != nil {
		log.Error().Caller().Err(err).Msgf("loadTile: latLng = %s -> error %s\n", ll.String(), err.Error())
		return nil, err
	}
	elevation, err := tile.GetElevation(ll)
	if err != nil {
		log.Error().Caller().Err(err).Msgf("GetElevation: latLng = %s -> error %s\n", ll.String(), err.Error())
		return nil, err
	}
	return append(point[:2], float64(elevation)), nil
}

// AddElevations returns geojson with added third coordinate (elevation)
// Param tileDir - directory of hgt-tiles
// Param geoJson - geojson for processing
// Param skipErrors - if false AddElevations use premature exit (on first bad point in geojson). if true all points will be process but bad point will not to be contains elevation coordinate
func AddElevations(tileDir string, geoJson *geojson.Geometry, skipErrors bool) (*geojson.Geometry, error) {
	switch geoJson.Type {
	case geojson.GeometryPoint:
		point, err := AddElevation(tileDir, geoJson.Point)
		if err != nil && !skipErrors {
			return nil, err
		}
		geoJson.Point = point
		return geoJson, nil
	case geojson.GeometryLineString:
		wg := &sync.WaitGroup{}
		wg.Add(len(geoJson.LineString))
		for i := range geoJson.LineString {
			go func(i int) {
				defer wg.Done()
				point, err := AddElevation(tileDir, geoJson.LineString[i])
				if err != nil && !skipErrors {
					log.Error().Caller().Err(err).Msg("")
				}
				geoJson.LineString[i] = point
			}(i)
		}
		wg.Wait()
		return geoJson, nil
	case geojson.GeometryMultiPoint:
		wg := &sync.WaitGroup{}
		wg.Add(len(geoJson.MultiPoint))
		for i := range geoJson.MultiPoint {
			go func(i int) {
				point, err := AddElevation(tileDir, geoJson.MultiPoint[i])
				if err != nil && !skipErrors {
					log.Error().Caller().Err(err).Msg("")
				}
				geoJson.MultiPoint[i] = point
			}(i)
		}
		wg.Wait()
		return geoJson, nil
	case geojson.GeometryPolygon:
		wg := &sync.WaitGroup{}
		for i := range geoJson.Polygon {
			wg.Add(len(geoJson.Polygon[i]))
			for j := range geoJson.Polygon[i] {
				go func(i, j int) {
					point, err := AddElevation(tileDir, geoJson.Polygon[i][j])
					if err != nil && !skipErrors {
						log.Error().Caller().Err(err).Msg("")
					}
					geoJson.Polygon[i][j] = point
				}(i, j)
			}
		}
		wg.Wait()
		return geoJson, nil
	case geojson.GeometryMultiLineString:
		wg := &sync.WaitGroup{}
		for i := range geoJson.MultiLineString {
			wg.Add(len(geoJson.MultiLineString[i]))
			for j, point := range geoJson.MultiLineString[i] {
				go func(i, j int) {
					point, err := AddElevation(tileDir, point)
					if err != nil && !skipErrors {
						log.Error().Caller().Err(err).Msg("")
					}
					geoJson.MultiLineString[i][j] = point
				}(i, j)
			}
		}
		wg.Wait()
		return geoJson, nil
	default:
		return geoJson, nil
	}
}
