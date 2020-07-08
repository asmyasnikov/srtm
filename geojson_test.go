package srtm

import (
	geojson "github.com/paulmach/go.geojson"
	"github.com/stretchr/testify/require"
	"math"
	"os"
	"path"
	"testing"
)

func TestAddElevations_Point(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	tileDir := path.Join(wd, "testdata")
	point, err := geojson.UnmarshalGeometry([]byte(`{"type":"Point","coordinates":[-65.92054637662613,-45.02475838113942]}`))
	require.NoError(t, err)
	point, err = AddElevations(tileDir, true, point, false)
	require.NoError(t, err)
	require.Equal(t, geojson.GeometryPoint, point.Type)
	require.Equal(t, 3, len(point.Point))
	require.Equal(t, 25, int(math.Round(point.Point[2])))
	b, err := point.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, `{"type":"Point","coordinates":[-65.92054637662613,-45.02475838113942,24.874129324015318]}`, string(b))
}

func TestAddElevations_LineString(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	tileDir := path.Join(wd, "testdata")
	lineString, err := geojson.UnmarshalGeometry([]byte(`{"type":"LineString","coordinates":[[-65.92054637662613,-45.02475838113942],[-65.92054637362613,-45.02475838114942],[-65.92053637662613,-45.02475835113942]]}`))
	require.NoError(t, err)
	require.Equal(t, 3, len(lineString.LineString))
	for _, point := range lineString.LineString {
		require.Equal(t, 2, len(point))
	}
	lineString, err = AddElevations(tileDir, true, lineString, false)
	require.NoError(t, err)
	require.Equal(t, geojson.GeometryLineString, lineString.Type)
	require.Equal(t, 3, len(lineString.LineString))
	for _, point := range lineString.LineString {
		require.Equal(t, 3, len(point))
	}
	b, err := lineString.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, `{"type":"LineString","coordinates":[[-65.92054637662613,-45.02475838113942,24.874129324015318],[-65.92054637362612,-45.02475838114942,24.87413069507827],[-65.92053637662613,-45.02475835113942,24.87891606292618]]}`, string(b))
}
