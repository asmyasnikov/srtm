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
	point, err := geojson.UnmarshalGeometry([]byte(`{"type":"Point","coordinates":[-65.978894751,-45.993612885]}`))
	require.NoError(t, err)
	point, err = AddElevations(tileDir, point, false)
	require.NoError(t, err)
	require.Equal(t, geojson.GeometryPoint, point.Type)
	require.Equal(t, 3, len(point.Point))
	require.Equal(t, 17, int(math.Round(point.Point[2])))
	b, err := point.MarshalJSON()
	require.Equal(t, `{"type":"Point","coordinates":[-65.978894751,-45.993612885,17.00000276401252]}`, string(b))
}

func TestAddElevations_LineString(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	tileDir := path.Join(wd, "testdata")
	lineString, err := geojson.UnmarshalGeometry([]byte(`{"type":"LineString","coordinates":[[-65.978894751,-45.993612885],[-65.978494751,-45.993614885],[-65.975894751,-45.993662885]]}`))
	require.NoError(t, err)
	require.Equal(t, 3, len(lineString.LineString))
	for _, point := range lineString.LineString {
		require.Equal(t, 2, len(point))
	}
	lineString, err = AddElevations(tileDir, lineString, false)
	require.NoError(t, err)
	require.Equal(t, geojson.GeometryLineString, lineString.Type)
	require.Equal(t, 3, len(lineString.LineString))
	for _, point := range lineString.LineString {
		require.Equal(t, 3, len(point))
	}
	b, err := lineString.MarshalJSON()
	require.Equal(t, `{"type":"LineString","coordinates":[[-65.978894751,-45.993612885,17.00000276401252],[-65.978494751,-45.993614885,18.42599987900539],[-65.975894751,-45.993662885,16]]}`, string(b))
}
