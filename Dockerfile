ARG GOARCH=amd64
ARG ARCH=amd64

FROM golang:latest AS build

RUN CGO_ENABLED=0 GOARCH=${GOARCH} go get -ldflags="-w -s" github.com/asmyasnikov/srtm/srtm-service

ARG ARCH

FROM --platform=linux/${ARCH} scratch

COPY --from=build /go/bin/srtm-service /srtm-service

ENV HTTP_PORT=80
ENV TILE_DIRECTORY=/data
ENV LRU_CACHE_SIZE=100
ENV STORE_IN_MEMORY=false

EXPOSE 80

ENTRYPOINT ["/srtm-service"]

CMD []
