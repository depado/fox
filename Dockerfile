# Build Step
FROM golang:1.25.3-alpine@sha256:aee43c3ccbf24fdffb7295693b6e33b21e01baec1b2a55acc351fde345e9ec34 as builder

# Dependencies
RUN apk update && apk add --no-cache upx make git
COPY --from=mwader/static-ffmpeg:7.1.1@sha256:11a44711684c0b9f754c047dcd64235b8b52deab251bd0e0a86f22faa160749c /ffmpeg /tmp/ffmpeg

# Source
WORKDIR $GOPATH/src/github.com/depado/fox
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify
COPY . .

# Build
RUN make tmp

# Final Step
FROM gcr.io/distroless/static@sha256:87bce11be0af225e4ca761c40babb06d6d559f5767fbf7dc3c47f0f1a466b92c
COPY --from=builder /tmp/fox /go/bin/fox
COPY --from=builder /tmp/ffmpeg /usr/bin/ffmpeg

VOLUME [ "/data" ]
WORKDIR /data
ENTRYPOINT ["/go/bin/fox"]
