FROM golang:latest AS build

WORKDIR /build

RUN CGO_ENABLED=0 go get -ldflags="-w -s" github.com/asmyasnikov/srtm/srtm-service

FROM scratch

COPY --from=build /go/bin/srtm-service /srtm-service

ENTRYPOINT ["/srtm-service"]

CMD []
