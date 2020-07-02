package srtm

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/pkg/errors"
)

var negativeDD = regexp.MustCompile(`S|W`)
var positiveDD = regexp.MustCompile(`N|E`)
var latitudeDD = regexp.MustCompile(`N|S`)

// ErrInvalidCoordDegrees is returned when latitude/longitude is
// unparsable or otherwise invalid
var ErrInvalidCoordDegrees = errors.New("invalid lat/lon degrees")

// LatLng represents a location
type LatLng struct {
	Latitude  float64
	Longitude float64
}

func (ll *LatLng) String() string {
	return fmt.Sprintf("[%0.7f, %0.7f]", ll.Latitude, ll.Longitude)
}

// dToDecimal accepts a direction-signed coordinate value (e.g. W|E or N|S prefix)
// and returns a positive or negative number instead
func dToDecimal(d string) (dd float64, err error) {
	makeNegative := false

	// make sure d is long enough so we can't get runtime error on string slicing
	if len(d) < 2 {
		return dd, errors.Wrap(ErrInvalidCoordDegrees, "too short, must contain direction and at least one digit")
	}

	dir := d[:1]

	// valid direction sign?
	if positiveDD.MatchString(d) {
	} else if negativeDD.MatchString(d) {
		makeNegative = true
	} else {
		return dd, errors.Wrapf(ErrInvalidCoordDegrees, "%s it not valid cardinal direction", dir)
	}

	i, err := strconv.Atoi(d[1:])
	if err != nil {
		return dd, errors.Wrap(ErrInvalidCoordDegrees, "could not convert to coord to int")
	}

	if i < 0 {
		return dd, errors.Wrapf(ErrInvalidCoordDegrees, "negative coord %f should is not valid in combination with direction,", dd)
	}

	if latitudeDD.MatchString(dir) {
		if i > 90 {
			return dd, errors.Wrap(ErrInvalidCoordDegrees, "latitude must be between 0 and 90")
		}
	} else {
		if i > 180 {
			return dd, errors.Wrap(ErrInvalidCoordDegrees, "longitude must be between 0 and 180")
		}
	}

	dd = float64(i)

	if makeNegative {
		dd = dd * -1
	}

	return dd, nil
}
