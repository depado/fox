# Build Step
FROM golang:1.25.6-alpine@sha256:98e6cffc31ccc44c7c15d83df1d69891efee8115a5bb7ede2bf30a38af3e3c92 as builder

# Dependencies
RUN apk update && apk add --no-cache upx make git
COPY --from=mwader/static-ffmpeg:8.0.1@sha256:252705ff88532fa338e7065c21792756552f8fe7c212f84bc503d3c340689594 /ffmpeg /tmp/ffmpeg

# Source
WORKDIR $GOPATH/src/github.com/depado/fox
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify
COPY . .

# Build
RUN make tmp

# Final Step
FROM gcr.io/distroless/static@sha256:cd64bec9cec257044ce3a8dd3620cf83b387920100332f2b041f19c4d2febf93
COPY --from=builder /tmp/fox /go/bin/fox
COPY --from=builder /tmp/ffmpeg /usr/bin/ffmpeg

VOLUME [ "/data" ]
WORKDIR /data
ENTRYPOINT ["/go/bin/fox"]
