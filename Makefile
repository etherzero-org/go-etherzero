# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: getz android ios getz-cross swarm evm all test clean
.PHONY: getz-linux getz-linux-386 getz-linux-amd64 getz-linux-mips64 getz-linux-mips64le
.PHONY: getz-linux-arm getz-linux-arm-5 getz-linux-arm-6 getz-linux-arm-7 getz-linux-arm64
.PHONY: getz-darwin getz-darwin-386 getz-darwin-amd64
.PHONY: getz-windows getz-windows-386 getz-windows-amd64

GOBIN = $(shell pwd)/build/bin
GO ?= latest

getz:
	build/env.sh go run build/ci.go install ./cmd/getz
	@echo "Done building."
	@echo "Run \"$(GOBIN)/getz\" to launch getz."

swarm:
	build/env.sh go run build/ci.go install ./cmd/swarm
	@echo "Done building."
	@echo "Run \"$(GOBIN)/swarm\" to launch swarm."

all:
	build/env.sh go run build/ci.go install

android:
	build/env.sh go run build/ci.go aar --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/getz.aar\" to use the library."

ios:
	build/env.sh go run build/ci.go xcode --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/Getz.framework\" to use the library."

test: all
	build/env.sh go run build/ci.go test

clean:
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

devtools:
	env GOBIN= go get -u golang.org/x/tools/cmd/stringer
	env GOBIN= go get -u github.com/jteeuwen/go-bindata/go-bindata
	env GOBIN= go get -u github.com/fjl/gencodec
	env GOBIN= go install ./cmd/abigen

# Cross Compilation Targets (xgo)

getz-cross: getz-linux getz-darwin getz-windows getz-android getz-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/getz-*

getz-linux: getz-linux-386 getz-linux-amd64 getz-linux-arm getz-linux-mips64 getz-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/getz-linux-*

getz-linux-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/getz
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/getz-linux-* | grep 386

getz-linux-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/getz
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/getz-linux-* | grep amd64

getz-linux-arm: getz-linux-arm-5 getz-linux-arm-6 getz-linux-arm-7 getz-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/getz-linux-* | grep arm

getz-linux-arm-5:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/getz
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/getz-linux-* | grep arm-5

getz-linux-arm-6:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/getz
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/getz-linux-* | grep arm-6

getz-linux-arm-7:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/getz
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/getz-linux-* | grep arm-7

getz-linux-arm64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/getz
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/getz-linux-* | grep arm64

getz-linux-mips:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/getz
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/getz-linux-* | grep mips

getz-linux-mipsle:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/getz
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/getz-linux-* | grep mipsle

getz-linux-mips64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/getz
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/getz-linux-* | grep mips64

getz-linux-mips64le:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/getz
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/getz-linux-* | grep mips64le

getz-darwin: getz-darwin-386 getz-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/getz-darwin-*

getz-darwin-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/getz
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/getz-darwin-* | grep 386

getz-darwin-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/getz
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/getz-darwin-* | grep amd64

getz-windows: getz-windows-386 getz-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/getz-windows-*

getz-windows-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/getz
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/getz-windows-* | grep 386

getz-windows-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/getz
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/getz-windows-* | grep amd64
