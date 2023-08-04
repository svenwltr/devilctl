FROM golang:1.20-alpine as builder

RUN apk add --no-cache git openssl

COPY . /build
RUN cd /build && ./buildutil

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Europe/Berlin /etc/localtime && \
    echo "Europe/Berlin" > /etc/timezone && \
    apk del tzdata

COPY --from=builder /build/dist/devilctl /usr/local/bin/

RUN adduser -D devilctl
USER devilctl
