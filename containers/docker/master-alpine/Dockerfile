FROM alpine:3.7

RUN \
  apk add --update go git make gcc musl-dev linux-headers ca-certificates && \
  git clone --depth 1 --branch release/1.8 https://github.com/etherzero-org/go-etherzero && \
  (cd go-etherzero && make geth) && \
  cp go-etherzero/build/bin/geth /geth && \
  cp go-etherzero/init.bin.1 /init.bin.1 && \
  cp go-etherzero/init.bin.2 /init.bin.2 && \
  cp go-etherzero/init.bin.3 /init.bin.3 && \
  apk del go git make gcc musl-dev linux-headers && \
  rm -rf /go-etherzero && rm -rf /var/cache/apk/*

EXPOSE 9646
EXPOSE 21212

ENTRYPOINT ["/geth"]
