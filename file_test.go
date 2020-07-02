package srtm

import (
	"encoding/binary"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

var testKeys = []string{
	"S46W066",
	"S46W067",
}

func TestReadFile(t *testing.T) {
	wd, _ := os.Getwd()
	for _, key := range testKeys {
		t.Run(key, func(t *testing.T) {
			tFileName, err := tilePath(path.Join(wd, "testdata"), key, LatLng{-46, -66})
			require.NoError(t, err)
			_, _, _, err = ReadFile(tFileName)
			require.NoError(t, err)
		})
	}

}

func TestBigEndingNegative(t *testing.T) {
	v := int16(binary.BigEndian.Uint16([]byte{byte(206), byte(180)}))
	require.Equal(t, int16(-12620), v)
}