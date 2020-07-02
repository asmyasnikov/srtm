package srtm

import (
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// ErrInvalidHGTFileName is returned when a HGT file name does not match
// pattern of a valid file which indicates lat/lon
var ErrInvalidHGTFileName = errors.New("invalid HGT file name")

var srtmParseName = regexp.MustCompile(`(N|S)(\d\d)(E|W)(\d\d\d)\.hgt(\.gz)?`)

// ReadFile is a helper func around Read that reads a SRTM file, decompressing
// if necessary, and returns  SRTM elevation data
func ReadFile(file string) (sw *LatLng, squareSize int, elevations []int16, err error) {
	f, err := os.Open(file)
	if err != nil {
		return sw, squareSize, elevations, err
	}
	defer f.Close()

	if strings.HasSuffix(file, ".gz") {
		rdr, err := gzip.NewReader(f)
		if err != nil {
			return sw, squareSize, elevations, err
		}
		defer rdr.Close()
		bytes, err := ioutil.ReadAll(rdr)
		if err != nil {
			return sw, squareSize, elevations, err
		}
		return Read(file, bytes)
	}
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return sw, squareSize, elevations, err
	}
	return Read(file, bytes)
}

// Read reads elevation for points from a SRTM file
func Read(fname string, bytes []byte) (sw *LatLng, squareSize int, elevations []int16, err error) {
	if len(bytes) == 12967201 * 2 {
		// 1 arcsecond
		squareSize = 3601
	} else if len(bytes) == 1442401 * 2 {
		// 3 arcseconds
		squareSize = 1201
	} else {
		return sw, squareSize, elevations, fmt.Errorf("hgt file cannot identified (only 1 arcsecond and 3 arcsecond supported, file size = %d)", len(bytes))
	}

	sw, err = southWest(fname)
	if err != nil {
		return sw, squareSize, elevations, errors.Wrap(err, "could not get corner coordinates from file name")
	}

	elevations = make([]int16, squareSize*squareSize)

	// Latitude
	for row := 0; row < squareSize; row++ {
		// Longitude
		for col := 0; col < squareSize; col++ {
			idx := row * squareSize + col
			elevations[idx] = int16(binary.BigEndian.Uint16(bytes[idx*2:idx*2+2]))
		}
	}

	return sw, squareSize, elevations, nil
}

// sw returns the southwest point contained in a HGT file.
// Coordinates in the file are relative to this point
func southWest(file string) (p *LatLng, err error) {
	fnameParts := srtmParseName.FindStringSubmatch(file)
	if fnameParts == nil {
		return p, ErrInvalidHGTFileName
	}

	swLatitude, err := dToDecimal(fnameParts[1] + fnameParts[2])
	if err != nil {
		return p, errors.Wrap(err, "could not get Latitude from file name")
	}
	swLongitude, err := dToDecimal(fnameParts[3] + fnameParts[4])
	if err != nil {
		return p, errors.Wrap(err, "could not get Longitude from file name")
	}

	return &LatLng{
		Latitude:  swLatitude,
		Longitude: swLongitude,
	}, err
}

// IsHGT returns true if fname appears to be a SRTM HGT file
func IsHGT(fname string) bool {
	if strings.HasSuffix(fname, ".hgt") {
		return true
	}

	if strings.HasSuffix(fname, ".hgt.gz") {
		return true
	}

	return false
}
