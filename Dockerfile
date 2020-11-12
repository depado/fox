# Build Step
FROM golang:1.15.4-alpine3.12 AS builder

# Dependencies
RUN apk update && apk add --no-cache upx
COPY --from=mwader/static-ffmpeg:4.3.1-2 /ffmpeg /tmp/ffmpeg
RUN upx /tmp/ffmpeg

# Source
WORKDIR $GOPATH/src/github.com/Depado/fox
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify
COPY . .

# Build
ARG build
ARG version
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=${version} -X main.Build=${build}" -o /tmp/fox
RUN upx /tmp/fox


# Final Step
FROM gcr.io/distroless/static
COPY --from=builder /tmp/fox /go/bin/fox
COPY --from=builder /tmp/ffmpeg /usr/bin/ffmpeg

VOLUME [ "/data" ]
WORKDIR /data
ENTRYPOINT ["/go/bin/fox"]
