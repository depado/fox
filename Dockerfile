# Build Step
FROM golang:1.15.4-alpine3.12 AS builder

# Prerequisites
RUN apk update && apk add --no-cache upx

# Dependencies
WORKDIR $GOPATH/src/github.com/Depado/fox
COPY . .
RUN go mod download
RUN go mod verify

# Build
ARG build
ARG version
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=${version} -X main.Build=${build}" -o /tmp/fox
# RUN upx --brute /tmp/fox

# Final step
FROM jrottenberg/ffmpeg:3.2-alpine

COPY --from=builder /tmp/fox /go/bin/fox
VOLUME [ "/data" ]
WORKDIR /data
ENTRYPOINT ["/go/bin/fox"]
