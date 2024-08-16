# Build Step
FROM golang:1.23-alpine AS builder

# Dependencies
RUN apk update && apk add --no-cache upx make git
COPY --from=mwader/static-ffmpeg:7.0.2 /ffmpeg /tmp/ffmpeg

# Source
WORKDIR $GOPATH/src/github.com/Depado/scarecrow
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify
COPY . .

# Build
RUN make tmp

# Final Step
FROM gcr.io/distroless/static
COPY --from=builder /tmp/fox /go/bin/fox
COPY --from=builder /tmp/ffmpeg /usr/bin/ffmpeg

VOLUME [ "/data" ]
WORKDIR /data
ENTRYPOINT ["/go/bin/fox"]
