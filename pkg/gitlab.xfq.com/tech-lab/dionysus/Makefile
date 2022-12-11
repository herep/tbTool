#Makefile


# make lint
include Makefile.basic.mk

GOPATH:=$(shell go env GOPATH)

GoVersion=$(shell go version)
BuildTime=$(shell date +%F-%Z/%T)
CommitID=$(shell git rev-parse HEAD)
LDFLAGS=-ldflags "-X 'gitlab.xfq.com/tech-lab/dionysus.GoVersion=${GoVersion}' \
-X gitlab.xfq.com/tech-lab/dionysus.BuildTime=${BuildTime} \
-X gitlab.xfq.com/tech-lab/dionysus.CommitID=${CommitID}"  \


.PHONY: mod-vendor
mod-vendor:
	go mod vendor

.PHONY: build-vendor
build-vendor:
	go mod vendor && go build ${LDFLAGS} -v -mod vendor -o dionysus example/example.go

.PHONY: build
build:
	go build ${LDFLAGS} -v -o dionysus example/example.go


.PHONY: release
release: build-vendor

.PHONY: test
test:
	go test -v ./... -cover

.PHONY: docker
docker:
	docker build . -t dionysus:latest