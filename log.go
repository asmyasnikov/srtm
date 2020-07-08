package srtm

import "github.com/rs/zerolog"

func init() {
	zerolog.TimeFieldFormat = "2006.01.02-15:04:05.000"
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}
