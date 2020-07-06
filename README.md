# srtm

[![GoDoc](https://godoc.org/github.com/asmyasnikov/srtm?status.svg)](https://godoc.org/github.com/asmyasnikov/srtm)
[![Go Report Card](https://goreportcard.com/badge/github.com/asmyasnikov/srtm)](https://goreportcard.com/report/github.com/asmyasnikov/srtm)
[![Build Status](https://travis-ci.org/asmyasnikov/srtm.png)](https://travis-ci.org/asmyasnikov/srtm)

Go library for reading [Shuttle Radar Topography Mission](https://en.wikipedia.org/wiki/Shuttle_Radar_Topography_Mission) (SRTM) HGT files

Based on [github.com/jda/srtm](https://github.com/jda/srtm) and inspired [geojson-elevation](https://github.com/perliedman/geojson-elevation) and [node-hgt](https://github.com/perliedman/node-hgt)

Written on pure golang. Golang realization:
 - use 1/2 of memory (instead nodejs elevation-service) in runtime 
 - golang slimmed docker image (bsed on busybox) have only 11MB (instead 161MB of slimmed nodeJs elevation-service)
 - 4676.27 trans/sec instead 1226.79 trans/sec for nodejs levation-service (tested with siege tool)

Support 1-arcsecond and 3-arcseconds hgt-tiles.

Provide web-service as elevation-service (like [github.com/asmyasnikov/elevation-service](https://github.com/asmyasnikov/elevation-service)) with allow CORS requests, auto-download zipped hgt-tiles from [imagico service](http://www.imagico.de/), unzipp and persist hgt-tiles in user-defined tile directory.

Install and usage:
```
# go get github.com/asmyasnikov/srtm/srtm-service
# mkdir data
# HTTP_PORT=80 TILE_DIRECTORY=./data LRU_CACHE_SIZE=4 srtm-service &
# curl -X POST -d '{"type":"LineString","coordinates":[[8.399786506567509,47.3439995300119],[8.401089653337102,47.34382901539513],[8.402392791687875,47.34365848600848],[8.403695921619205,47.343487941852196],[8.404999043130463,47.343317382926415],[8.406302156221027,47.34314680923133],[8.407605260890275,47.34297622076714],[8.408908357137577,47.342805617534005],[8.410211444962314,47.342634999532144],[8.41151452436386,47.3424643667617],[8.412817595341588,47.34229371922288],[8.414120657894879,47.34212305691586],[8.415423712023102,47.34195237984083],[8.41672675772564,47.34178168799798],[8.418029795001864,47.34161098138748],[8.419332823851153,47.34144026000951],[8.42063584427288,47.34126952386427],[8.421938856266422,47.34109877295194],[8.423241859831155,47.34092800727269],[8.424544854966458,47.3407572268267],[8.425847841671704,47.340586431614206],[8.427150819946267,47.34041562163533],[8.428453789789527,47.340244796890275],[8.42975675120086,47.340073957379225],[8.43105970417964,47.33990310310238],[8.432362648725245,47.33973223405991],[8.43366558483705,47.33956135025199],[8.434968512514432,47.339390451678824],[8.436271431756767,47.33921953834058],[8.437574342563435,47.33904861023746],[8.438877244933805,47.338877667369644],[8.440180138867259,47.338706709737295],[8.441483024363173,47.338535737340614],[8.44278590142092,47.33836475017978],[8.444088770039883,47.33819374825498],[8.445391630219433,47.3380227315664],[8.446694481958948,47.3378517001142],[8.447997325257806,47.337680653898616],[8.449300160115381,47.33750959291979],[8.450602986531052,47.33733851717791],[8.451905804504195,47.33716742667317],[8.453208614034189,47.336996321405756],[8.454511415120406,47.33682520137584],[8.455814207762229,47.33665406658363],[8.45711699195903,47.33648291702928],[8.458419767710186,47.33631175271298],[8.459722535015079,47.33614057363493],[8.46102529387308,47.33596937979531],[8.46232804428357,47.3357981711943],[8.463630786245924,47.335626947832075],[8.463638463275133,47.3356259387696]]}' http://localhost/ 
```
Requset return geojson response with third coordinate (elevation) in each point:
```json
{"type":"LineString","coordinates":[[8.399786506567509,47.3439995300119,630.833146255931],[8.401089653337102,47.34382901539513,631.1311413898052],[8.402392791687875,47.34365848600848,627.2093291109096],[8.403695921619205,47.343487941852196,618.6073505976871],[8.404999043130463,47.343317382926415,607.8155065555864],[8.406302156221027,47.34314680923133,592.85297273949],[8.407605260890275,47.34297622076714,586.8342281017506],[8.408908357137577,47.342805617534005,584.8388532151023],[8.410211444962314,47.342634999532144,584.5515300526016],[8.41151452436386,47.3424643667617,588.9304513141012],[8.412817595341588,47.34229371922288,591.2375981452501],[8.414120657894879,47.34212305691586,591.4417363870025],[8.415423712023102,47.34195237984083,574.745089862652],[8.41672675772564,47.34178168799798,561.1824938129538],[8.418029795001864,47.34161098138748,589.7004457467012],[8.419332823851153,47.34144026000951,609.2399016173551],[8.42063584427288,47.34126952386427,612.9586671574217],[8.421938856266422,47.34109877295194,602.9499045025316],[8.423241859831155,47.34092800727269,595.8188245240952],[8.424544854966458,47.3407572268267,602.316197596573],[8.425847841671704,47.340586431614206,607.6214264931714],[8.427150819946267,47.34041562163533,591.2713476941213],[8.428453789789527,47.340244796890275,585.459610043686],[8.42975675120086,47.340073957379225,582.7094835022643],[8.43105970417964,47.33990310310238,572.850030606854],[8.432362648725245,47.33973223405991,559.7566129760144],[8.43366558483705,47.33956135025199,548.9537315824814],[8.434968512514432,47.339390451678824,541.1714336414659],[8.436271431756767,47.33921953834058,536.5196463999931],[8.437574342563435,47.33904861023746,533.2833354300903],[8.438877244933805,47.338877667369644,536.2780068125097],[8.440180138867259,47.338706709737295,544.0218545442497],[8.441483024363173,47.338535737340614,553.0026456995458],[8.44278590142092,47.33836475017978,562.563505115433],[8.444088770039883,47.33819374825498,573.8908859235255],[8.445391630219433,47.3380227315664,587.1385537056937],[8.446694481958948,47.3378517001142,589.8809166543787],[8.447997325257806,47.337680653898616,591.991887877906],[8.449300160115381,47.33750959291979,598.3422610065583],[8.450602986531052,47.33733851717791,618.3481418074465],[8.451905804504195,47.33716742667317,631.0063391815225],[8.453208614034189,47.336996321405756,629.063756319398],[8.454511415120406,47.33682520137584,622.2752814297355],[8.455814207762229,47.33665406658363,614.3123968654625],[8.45711699195903,47.33648291702928,582.1021102875286],[8.458419767710186,47.33631175271298,546.5127095572064],[8.459722535015079,47.33614057363493,529.5937443126201],[8.46102529387308,47.33596937979531,526.0294904980549],[8.46232804428357,47.3357981711943,525.3704991525899],[8.463630786245924,47.335626947832075,523.731458059482],[8.463638463275133,47.3356259387696,523.7249705691114]]}
``` 

Usage from sources
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
