FROM golang:latest AS build

WORKDIR /build

RUN CGO_ENABLED=0 go get -ldflags="-w -s" github.com/asmyasnikov/srtm/srtm-service

FROM amd64/busybox:glibc

COPY --from=build /go/bin/srtm-service /usr/bin/srtm-service

ENTRYPOINT ["/usr/bin/srtm-service"]

CMD []
