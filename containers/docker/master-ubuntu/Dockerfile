FROM ubuntu:xenial

RUN \
  apt-get update && apt-get upgrade -q -y && \
  apt-get install -y --no-install-recommends golang git make gcc libc-dev ca-certificates && \
  git clone --depth 1 --branch release/1.7 https://github.com/ethzero/go-ethzero && \
  (cd go-ethzero && make getz) && \
  cp go-ethzero/build/bin/getz /getz && \
  apt-get remove -y golang git make gcc libc-dev && apt autoremove -y && apt-get clean && \
  rm -rf /go-ethzero

EXPOSE 9646
EXPOSE 21212

ENTRYPOINT ["/getz"]
