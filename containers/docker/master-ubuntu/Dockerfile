FROM ubuntu:xenial

ENV PATH=/usr/lib/go-1.9/bin:$PATH

RUN \
  apt-get update && apt-get upgrade -q -y && \
  apt-get install -y --no-install-recommends golang-1.9 git make gcc libc-dev ca-certificates && \
  git clone --depth 1 --branch release/1.8 https://github.com/etherzero-org/go-etherzero && \
  (cd go-etherzero && make geth) && \
  cp go-etherzero/build/bin/geth /geth && \
  cp go-etherzero/init.bin.1 /init.bin.1 && \
  cp go-etherzero/init.bin.2 /init.bin.2 && \
  cp go-etherzero/init.bin.3 /init.bin.3 && \
  apt-get remove -y golang-1.9 git make gcc libc-dev && apt autoremove -y && apt-get clean && \
  rm -rf /go-etherzero

EXPOSE 9646
EXPOSE 21212

ENTRYPOINT ["/geth"]
