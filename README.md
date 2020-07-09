# srtm

[![GoDoc](https://godoc.org/github.com/asmyasnikov/srtm?status.svg)](https://godoc.org/github.com/asmyasnikov/srtm)
[![Go Report Card](https://goreportcard.com/badge/github.com/asmyasnikov/srtm)](https://goreportcard.com/report/github.com/asmyasnikov/srtm)
[![Build Status](https://travis-ci.org/asmyasnikov/srtm.svg)](https://travis-ci.org/asmyasnikov/srtm)

Go library for reading [Shuttle Radar Topography Mission](https://en.wikipedia.org/wiki/Shuttle_Radar_Topography_Mission) (SRTM) HGT files

Written on pure golang. Based on [github.com/jda/srtm](https://github.com/jda/srtm) and inspired [geojson-elevation](https://github.com/perliedman/geojson-elevation) and [node-hgt](https://github.com/perliedman/node-hgt)

Compare testing results (tested on Intel Core i3-7100, 8GB memory, SSD, elevation/srtm services run inside docker with port forwarding). 

|                                                      | Memory usage at start, MB | Memory usage active phase, MB | docker slimmed image, MB | rps (siege trans/sec) |
|------------------------------------------------------|---------------------------|-------------------------------|--------------------------|-----------------------|
| node.js [elevation-service](https://github.com/asmyasnikov/elevation-service) | 40.44 | 77.41 | 161 | 837.61 |
| golang [srtm-service](github.com/asmyasnikov/srtm/srtm-service/) with env `STORE_IN_MEMORY=true` | 1.76 | 49.09 | 11  | 3214.50 |
| golang [srtm-service](github.com/asmyasnikov/srtm/srtm-service/) with env `STORE_IN_MEMORY=false` | 1.76 | 17.22 | 11 | 2816.02 |

Siege run from command
```
siege -t 5S -c 500 --content-type "application/json" 'http://localhost:18081/ POST {"type":"LineString","coordinates":[[8.399786506567509,47.3439995300119],[8.401089653337102,47.34382901539513],[8.402392791687875,47.34365848600848],[8.403695921619205,47.343487941852196],[8.404999043130463,47.343317382926415],[8.406302156221027,47.34314680923133],[8.407605260890275,47.34297622076714],[8.408908357137577,47.342805617534005],[8.410211444962314,47.342634999532144],[8.41151452436386,47.3424643667617],[8.412817595341588,47.34229371922288],[8.414120657894879,47.34212305691586],[8.415423712023102,47.34195237984083],[8.41672675772564,47.34178168799798],[8.418029795001864,47.34161098138748],[8.419332823851153,47.34144026000951],[8.42063584427288,47.34126952386427],[8.421938856266422,47.34109877295194],[8.423241859831155,47.34092800727269],[8.424544854966458,47.3407572268267],[8.425847841671704,47.340586431614206],[8.427150819946267,47.34041562163533],[8.428453789789527,47.340244796890275],[8.42975675120086,47.340073957379225],[8.43105970417964,47.33990310310238],[8.432362648725245,47.33973223405991],[8.43366558483705,47.33956135025199],[8.434968512514432,47.339390451678824],[8.436271431756767,47.33921953834058],[8.437574342563435,47.33904861023746],[8.438877244933805,47.338877667369644],[8.440180138867259,47.338706709737295],[8.441483024363173,47.338535737340614],[8.44278590142092,47.33836475017978],[8.444088770039883,47.33819374825498],[8.445391630219433,47.3380227315664],[8.446694481958948,47.3378517001142],[8.447997325257806,47.337680653898616],[8.449300160115381,47.33750959291979],[8.450602986531052,47.33733851717791],[8.451905804504195,47.33716742667317],[8.453208614034189,47.336996321405756],[8.454511415120406,47.33682520137584],[8.455814207762229,47.33665406658363],[8.45711699195903,47.33648291702928],[8.458419767710186,47.33631175271298],[8.459722535015079,47.33614057363493],[8.46102529387308,47.33596937979531],[8.46232804428357,47.3357981711943],[8.463630786245924,47.335626947832075],[8.463638463275133,47.3356259387696]]}'
```

Support 1-arcsecond and 3-arcseconds hgt-tiles.

Provide web-service as elevation-service (like [github.com/asmyasnikov/elevation-service](https://github.com/asmyasnikov/elevation-service)) with allow CORS requests, auto-download zipped hgt-tiles from [imagico service](http://www.imagico.de/), unzipp and persist hgt-tiles in user-defined tile directory.


Environment variables:
 - `HTTP_PORT` - http port of web-service (default 80)
 - `TILE_DIRECTORY` - directory of hgt tiles (default `./data/`)
 - `LRU_CACHE_SIZE` - LRU cache size (default 1000)
 - `STORE_IN_MEMORY` - boolean flag. If `false` hgt tiles not preliminary reading into memory, read few bytes at AddElevation phase. if `true` - all contents read preliminary into memory and store for future usage. (default `true`)
 - `LOG_LEVEL` - logging level 

Install and usage:
 - from sources 
```
# go get github.com/asmyasnikov/srtm/srtm-service
# mkdir data
# HTTP_PORT=80 TILE_DIRECTORY=./data LRU_CACHE_SIZE=1000 STORE_IN_MEMORY=false srtm-service &
```
 - from docker
```
# mkdir data
# docker run -itd --rm -v $(pwd)/data/:/data/ -p 80:80 -e HTTP_PORT=80 -e TILE_DIRECTORY=/data -e LRU_CACHE_SIZE=1000 -e STORE_IN_MEMORY=false -e LOG_LEVEL=debug amyasnikov/srtm-service:latest
```
 - in sources
```go
package main

import (
	"github.com/asmyasnikov/srtm"
	"log"
)

func main() {
    point, err := srtm.AddElevation("./data/", []float64{
        8.399786506567509, // longitude
        47.3439995300119, // latitude
    })
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Lat: %.7f, Lng: %.7f, Elevation: %.1f", point[1], point[0], point[2])
}
```
