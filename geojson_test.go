package srtm

import (
	geojson "github.com/paulmach/go.geojson"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"runtime"
	"testing"
)

var lineString *geojson.Geometry

func init() {
	coordinates := [][]float64{
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
		{-67.0 + rand.Float64(), -46.0 + rand.Float64()},
	}
	lineString = geojson.NewLineStringGeometry(coordinates)
	zerolog.SetGlobalLevel(zerolog.NoLevel)
}

func TestAddElevations_Point(t *testing.T) {
	data, err := New(1, "testdata", -1)
	require.NoError(t, err)
	defer data.Destroy()
	point, err := geojson.UnmarshalGeometry([]byte(`{"type":"Point","coordinates":[-65.92054637662613,-45.02475838113942]}`))
	require.NoError(t, err)
	err = data.AddElevations(point, false)
	require.NoError(t, err)
	require.Equal(t, geojson.GeometryPoint, point.Type)
	require.Equal(t, 3, len(point.Point))
	require.Equal(t, 25, int(math.Round(point.Point[2])))
	b, err := point.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, `{"type":"Point","coordinates":[-65.92054637662613,-45.02475838113942,24.874129324015318]}`, string(b))
}

func TestAddElevations_LineString(t *testing.T) {
	data, err := New(1, "testdata", -1)
	require.NoError(t, err)
	defer data.Destroy()
	lineString, err := geojson.UnmarshalGeometry([]byte(`{"type":"LineString","coordinates":[[-65.92054637662613,-45.02475838113942],[-65.92054637362613,-45.02475838114942],[-65.92053637662613,-45.02475835113942]]}`))
	require.NoError(t, err)
	require.Equal(t, 3, len(lineString.LineString))
	for _, point := range lineString.LineString {
		require.Equal(t, 2, len(point))
	}
	err = data.AddElevations(lineString, false)
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

func TestAddElevations_LineString_Rand(t *testing.T) {
	data, err := New(1, "testdata", -1)
	require.NoError(t, err)
	defer data.Destroy()
	require.NoError(t, data.AddElevations(lineString, false))
}

func Benchmark_process2_1(b *testing.B) {
	data, err := New(1, "testdata", -1)
	require.NoError(b, err)
	defer data.Destroy()
	data.process2(lineString.LineString, 1)
}

func Benchmark_process2_4(b *testing.B) {
	data, err := New(1, "testdata", -1)
	require.NoError(b, err)
	defer data.Destroy()
	data.process2(lineString.LineString, 4)
}

func Benchmark_process2_8(b *testing.B) {
	data, err := New(1, "testdata", -1)
	require.NoError(b, err)
	defer data.Destroy()
	data.process2(lineString.LineString, 8)
}

func Benchmark_process2_16(b *testing.B) {
	data, err := New(1, "testdata", -1)
	require.NoError(b, err)
	defer data.Destroy()
	data.process2(lineString.LineString, 16)
}

func Benchmark_process2_32(b *testing.B) {
	data, err := New(1, "testdata", -1)
	require.NoError(b, err)
	defer data.Destroy()
	data.process2(lineString.LineString, 32)
}

func Benchmark_process2_NCPU(b *testing.B) {
	data, err := New(1, "testdata", -1)
	require.NoError(b, err)
	defer data.Destroy()
	data.process2(lineString.LineString, runtime.NumCPU())
}