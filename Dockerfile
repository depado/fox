# Build Step
FROM golang:1.26.4-alpine@sha256:a6a091eac01ceac4b97496fe2957a49b6cdd83365337d5f46f6f73710424e805 as builder

# Dependencies
RUN apk update && apk add --no-cache upx make git
COPY --from=mwader/static-ffmpeg:8.1.1@sha256:735f84b905e00d5c618b667f0b053f83b1096f5fc404c607e6134bf2275a0e0a /ffmpeg /tmp/ffmpeg

# Source
WORKDIR $GOPATH/src/github.com/depado/fox
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify
COPY . .

# Build
RUN make tmp

# Final Step
FROM gcr.io/distroless/static@sha256:3592aa8171c77482f62bbc4164e6a2d141c6122554ace66e5cc910cadb961ff0
COPY --from=builder /tmp/fox /go/bin/fox
COPY --from=builder /tmp/ffmpeg /usr/bin/ffmpeg

VOLUME [ "/data" ]
WORKDIR /data
ENTRYPOINT ["/go/bin/fox"]
