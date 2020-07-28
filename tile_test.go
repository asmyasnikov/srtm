package srtm

import (
	"github.com/stretchr/testify/require"
	"math"
	"testing"
)

func TestGetElevation(t *testing.T) {
	data, err := New(1, "testdata", -1)
	require.NoError(t, err)
	defer data.Destroy()
	tile, err := data.loadTile(
		LatLng{
			Latitude:  -45.55457,
			Longitude: -65.23555,
		})
	require.NoError(t, err)
	require.Equal(t, 3601, tile.size)
	require.Equal(t, 3601*3601, len(tile.elevations))
	require.Equal(t, (&LatLng{
		Latitude:  -46,
		Longitude: -66,
	}).String(), tile.sw.String())
	e, err := tile.GetElevation(LatLng{
		Latitude:  -45.02475838113942,
		Longitude: -65.92054637662613,
	})
	require.NoError(t, err)
	require.Equal(t, 25, int(math.Round(e)))
}
