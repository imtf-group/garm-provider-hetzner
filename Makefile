SHELL := bash

ROOTDIR=$(dir $(abspath $(lastword $(MAKEFILE_LIST))))
GO ?= go
IMAGE_TAG = garm-provider-build
GITHUB_REF_NAME ?= latest

USER_ID=$(shell ((docker --version | grep -q podman) && echo "0" || id -u))
USER_GROUP=$(shell ((docker --version | grep -q podman) && echo "0" || id -g))
GARM_PROVIDER_NAME := garm-provider-hetzner

default: test build

clean:
	@rm -rf ./build ./vendor ./release ./$(GARM_PROVIDER_NAME)

test: install-lint-deps run-tests lint

release: build-static make-release

build:
	@$(GO) build .

build-static:
	go mod vendor
	docker build --tag $(IMAGE_TAG) .
	mkdir -p build
	docker run --rm -e VERSION=$(GITHUB_REF_NAME) -e GARM_PROVIDER_NAME=$(GARM_PROVIDER_NAME) -e USER_ID=$(USER_ID) -e USER_GROUP=$(USER_GROUP) -v $(PWD)/build:/build/output:z -v $(PWD):/build/$(GARM_PROVIDER_NAME):z $(IMAGE_TAG) /build-static.sh

install-lint-deps:
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

lint:
	@golangci-lint run --timeout=8m --build-tags testing

run-tests:
	$(GO) test -v ./... -coverprofile=cover.out

coverage:
	@$(GO) tool cover -func=cover.out

fmt:
	@$(GO) fmt $$($(GO) list ./...)

make-release:
	./scripts/make-release.sh
