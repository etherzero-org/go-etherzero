.PHONY: build cluster test

default: build

build:
	go build -o bin/swarm    github.com/ethzero/go-ethzero/cmd/swarm
	go build -o bin/geth     github.com/ethzero/go-ethzero/cmd/geth
	go build -o bin/bootnode github.com/ethzero/go-ethzero/cmd/bootnode

cluster: build
	scripts/boot-cluster.sh

test:
	go test -v github.com/ethzero/go-ethzero/swarm/...
