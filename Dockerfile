# Build Geth in a stock Go builder container
FROM golang:1.10-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers git 

ADD . /go-ethereum
RUN cd /go-ethereum && make geth

# Pull Geth into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /go-ethereum/build/bin/geth /usr/local/bin/
COPY init.bin.1 /root/
COPY init.bin.2 /root/
COPY init.bin.3 /root/
WORKDIR /root

EXPOSE 9646 8546 21212 21212/udp 21213/udp
#ENTRYPOINT ["geth"]
ENTRYPOINT ["geth","--rpc"] 
#,"--rpcaddr 0.0.0.0","--rpcport 9646" ,"--rpcvhosts *", "--rpccorsdomain *"]
