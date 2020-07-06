package srtm

import (
	"github.com/stretchr/testify/require"
	"math"
	"os"
	"path"
	"testing"
)

func TestGetElevation(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	tile, err := loadTile(
		path.Join(wd, "testdata"), LatLng{
			Latitude:  -45.55457,
			Longitude: -66.23555,
		})
	require.NoError(t, err)
	require.Equal(t, 3601, tile.size)
	require.Equal(t, func() int {
		if storeInMemoryMode() {
			return 3601 * 3601
		}
		return 0
	}(), len(tile.elevations))
	require.Equal(t, (&LatLng{
		Latitude:  -46,
		Longitude: -67,
	}).String(), tile.sw.String())
	e, err := tile.GetElevation(LatLng{
		Latitude:  -45.02475838113942,
		Longitude: -66.92054637662613,
	})
	require.NoError(t, err)
	require.Equal(t, 260, int(math.Round(e)))
}
