package srtm

import (
	"github.com/stretchr/testify/require"
	"math"
	"os"
	"path"
	"testing"
)

func TestLoadTile(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	tile, err := loadTile(
		path.Join(wd, "testdata"), LatLng{
		Latitude:  -45.5547457,
		Longitude: -65.2352355,
	})
	require.NoError(t, err)
	require.Equal(t, 3601, tile.size)
	require.Equal(t, 3601*3601, len(tile.elevations))
	require.Equal(t, (&LatLng{
		Latitude: -46,
		Longitude: -66,
	}).String(), tile.sw.String())
	e, err := tile.getElevation(LatLng{
		Latitude: -46.0 + 1.0 / 3601.0 * 23,
		Longitude: -66.0 + 1.0 / 3601.0 * 76,
	})
	require.NoError(t, err)
	require.Equal(t, 17, int(math.Round(e)))
}
