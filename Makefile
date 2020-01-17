APP=$(shell basename "$(PWD)")
RELEASE=$(shell git describe --abbrev=0 --tags)
PLATFORMS := linux/amd64 windows/amd64 darwin/amd64
temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))
# GOOS?=linux
# GOARCH?=amd64
DOCKER_REGISTRY?=registry.example.com
COMMIT=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

.PHONY: help build clean dep image push

all: help

build: clean dep $(PLATFORMS)

$(PLATFORMS): clean dep
	CGO_ENABLED=0 GOOS=$(os) GOARCH=$(arch) go build \
		-ldflags "-w -s -X ${APP}/pkg/version.version=${RELEASE} -X ${APP}/pkg/version.commit=${COMMIT} -X ${APP}/pkg/version.buildTime=${BUILD_TIME}" \
		-o bin/${APP}/${RELEASE}/${APP}_${RELEASE}_$(os)_$(arch)

clean:
	rm -rf bin/${APP}/${RELEASE}

dep:
	go mod download

image:
	docker build --build-arg "APP=${APP}" --build-arg "RELEASE=${RELEASE}" --build-arg "COMMIT=${COMMIT}" --tag "${DOCKER_REGISTRY}/library/${APP}:${RELEASE}" $(PWD)
	docker tag "${DOCKER_REGISTRY}/library/${APP}:${RELEASE}" "${DOCKER_REGISTRY}/library/${APP}:latest"

push:
	docker push "${DOCKER_REGISTRY}/library/${APP}:${RELEASE}"
	docker push "${DOCKER_REGISTRY}/library/${APP}:latest"

help: Makefile
	@echo
	@echo " Choose a command run in "$(APP)":"
	@echo
	@echo build
	@echo clean
	@echo dep
	@echo image
	@echo push
	@echo
