package srtm

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestImagicoResponseParse(t *testing.T) {
	b := `[{"name":"J40.zip","title":"","link":"http:\/\/www.viewfinderpanoramas.org\/dem3\/J40.zip","date":"2012-04-15","lon_start":"54","lat_start":"36","lon_end":"60","lat_end":"40","type":"2"}]`
	url, err := parse(bytes.NewBufferString(b))
	require.NoError(t, err)
	require.Equal(t, "http://www.viewfinderpanoramas.org/dem3/J40.zip", url)
}
