FROM alpine:3.5

RUN \
  apk add --update go git make gcc musl-dev linux-headers ca-certificates && \
  git clone --depth 1 --branch release/1.7 https://github.com/ethzero/go-ethzero && \
  (cd go-ethzero && make getz) && \
  cp go-ethzero/build/bin/getz /getz && \
  apk del go git make gcc musl-dev linux-headers && \
  rm -rf /go-ethzero && rm -rf /var/cache/apk/*

EXPOSE 9646
EXPOSE 21212

ENTRYPOINT ["/getz"]
