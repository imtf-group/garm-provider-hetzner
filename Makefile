SHELL := bash

GO ?= go

default: test build

build:
	@$(GO) build .

test:
	$(GO) test -v ./... -coverprofile=cover.out

coverage:
	@$(GO) tool cover -func=cover.out

fmt:
	@$(GO) fmt $$($(GO) list ./...)
