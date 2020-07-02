package srtm

import (
	geojson "github.com/paulmach/go.geojson"
)

func addElevation(tileDir string, point []float64) ([]float64, error) {
	ll := LatLng{
		Latitude:  point[1],
		Longitude: point[0],
	}
	tile, err := loadTile(tileDir, ll)
	if err != nil {
		return nil, err
	}
	elevation, err := tile.getElevation(ll)
	if err != nil {
		return nil, err
	}
	return append(point, float64(elevation)), nil
}

func AddElevations(tileDir string, geoJson *geojson.Geometry, skipErrors bool) (*geojson.Geometry, error) {
	switch geoJson.Type {
	case geojson.GeometryPoint:
		point, err := addElevation(tileDir, geoJson.Point)
		if err != nil && !skipErrors {
			return nil, err
		}
		geoJson.Point = point
		return geoJson, nil
	case geojson.GeometryLineString:
		for i, point := range geoJson.LineString {
			point, err := addElevation(tileDir, point)
			if err != nil && !skipErrors {
				return nil, err
			}
			geoJson.LineString[i] = point
		}
		return geoJson, nil
	case geojson.GeometryMultiPoint:
		for i, point := range geoJson.MultiPoint {
			point, err := addElevation(tileDir, point)
			if err != nil && !skipErrors {
				return nil, err
			}
			geoJson.MultiPoint[i] = point
		}
		return geoJson, nil
	case geojson.GeometryPolygon:
		for i := range geoJson.Polygon {
			for j, point := range geoJson.Polygon[i] {
				point, err := addElevation(tileDir, point)
				if err != nil && !skipErrors {
					return nil, err
				}
				geoJson.Polygon[i][j] = point
			}
		}
		return geoJson, nil
	case geojson.GeometryMultiLineString:
		for i := range geoJson.MultiLineString {
			for j, point := range geoJson.MultiLineString[i] {
				point, err := addElevation(tileDir, point)
				if err != nil && !skipErrors {
					return nil, err
				}
				geoJson.MultiLineString[i][j] = point
			}
		}
		return geoJson, nil
	default:
		return geoJson, nil
	}
}
