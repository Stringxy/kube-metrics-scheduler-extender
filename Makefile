# Manage platform and builders
PLATFORMS ?= linux/amd64,linux/arm64
BUILDER ?= docker
IMG ?=http://192.168.112.150:5001/kube-metrics-scheduler-extender:v1


.PHONY: build
build:
	go build -o bin/extender main.go

build-image:
	${BUILDER} buildx build --push --platform ${PLATFORMS} -t ${IMG} .