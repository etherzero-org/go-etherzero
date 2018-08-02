# Build Geth in a stock Go builder container
FROM golang:1.10-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers

ADD . /go-etherzero
RUN cd /go-etherzero && make all

# Pull all binaries into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /go-etherzero/build/bin/* /usr/local/bin/

EXPOSE 9646 8546 21212 21212/udp 21213/udp
