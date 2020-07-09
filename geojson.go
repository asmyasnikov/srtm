package srtm

import (
	geojson "github.com/paulmach/go.geojson"
	"github.com/rs/zerolog/log"
	"sync"
)

// AddElevation returns point with 3 coordinates: [longitude, latitude, elevation]
// Param tileDir - directory of hgt-tiles
// Param point - [longitude, latitude]
func AddElevation(point []float64) ([]float64, error) {
	ll := LatLng{
		Latitude:  point[1],
		Longitude: point[0],
	}
	tile, err := loadTile(ll)
	if err != nil {
		log.Error().Caller().Err(err).Msgf("loadTile: latLng = %s -> error %s", ll.String(), err.Error())
		return nil, err
	}
	elevation, err := tile.GetElevation(ll)
	if err != nil {
		log.Error().Caller().Err(err).Msgf("GetElevation: latLng = %s -> error %s", ll.String(), err.Error())
		return nil, err
	}
	return append(point[:2], float64(elevation)), nil
}

func processorLineString(wg *sync.WaitGroup, geoJson *geojson.Geometry, i int, skipErrors bool) {
	point, err := AddElevation(geoJson.LineString[i])
	if err != nil && !skipErrors {
		log.Error().Caller().Err(err).Msg("")
	}
	geoJson.LineString[i] = point
	if wg != nil {
		wg.Done()
	}
}

// AddElevations returns geojson with added third coordinate (elevation)
// Param tileDir - directory of hgt-tiles
// Param geoJson - geojson for processing
// Param skipErrors - if false AddElevations use premature exit (on first bad point in geojson). if true all points will be process but bad point will not to be contains elevation coordinate
func AddElevations(geoJson *geojson.Geometry, skipErrors bool) (*geojson.Geometry, error) {
	var wg *sync.WaitGroup
	if parallel {
		wg = &sync.WaitGroup{}
	}
	switch geoJson.Type {
	case geojson.GeometryPoint:
		point, err := AddElevation(geoJson.Point)
		if err != nil && !skipErrors {
			return nil, err
		}
		geoJson.Point = point
		return geoJson, nil
	case geojson.GeometryLineString:
		if parallel {
			wg.Add(len(geoJson.LineString))
		}
		for i := range geoJson.LineString {
			if parallel {
				go processorLineString(wg, geoJson, i, skipErrors)
			} else {
				processorLineString(nil, geoJson, i, skipErrors)
			}
		}
	case geojson.GeometryMultiPoint:
		for i := range geoJson.MultiPoint {
			point, err := AddElevation(geoJson.MultiPoint[i])
			if err != nil && !skipErrors {
				log.Error().Caller().Err(err).Msg("")
			}
			geoJson.MultiPoint[i] = point
		}
	case geojson.GeometryPolygon:
		for i := range geoJson.Polygon {
			for j := range geoJson.Polygon[i] {
				point, err := AddElevation(geoJson.Polygon[i][j])
				if err != nil && !skipErrors {
					log.Error().Caller().Err(err).Msg("")
				}
				geoJson.Polygon[i][j] = point
			}
		}
	case geojson.GeometryMultiLineString:
		for i := range geoJson.MultiLineString {
			for j := range geoJson.MultiLineString[i] {
				point, err := AddElevation(geoJson.MultiLineString[i][j])
				if err != nil && !skipErrors {
					log.Error().Caller().Err(err).Msg("")
				}
				geoJson.MultiLineString[i][j] = point
			}
		}
	}
	if parallel {
		wg.Wait()
	}
	return geoJson, nil
}
