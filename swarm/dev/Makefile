.PHONY: build cluster test

default: build

build:
	go build -o bin/swarm    github.com/etherzero/go-etherzero/cmd/swarm
	go build -o bin/geth     github.com/etherzero/go-etherzero/cmd/geth
	go build -o bin/bootnode github.com/etherzero/go-etherzero/cmd/bootnode

cluster: build
	scripts/boot-cluster.sh

test:
	go test -v github.com/etherzero/go-etherzero/swarm/...
