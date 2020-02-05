GO_BIN := $(shell which go)
GO_LINT := $(shell which golint)
UPX := $(shell which upx)
PROJECT_NAME := "blcheck"
PKG := "github.com/teqneers/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/...)
GO_FILES := $(shell find . -name '*.go' | grep -v _test.go)

.PHONY: all dep build clean test race lint buildmacos buildlinux compressmacos compresslinux release

all: release

lint: ## Lint the files
	${GO_LINT} -set_exit_status ${PKG_LIST}

test: ## Run unittests
	${GO_BIN} test -short ${PKG_LIST}

race: ## Run race tests
	${GO_BIN} test -race -short ${PKG_LIST}

dep: ## Get the dependencies
	${GO_BIN} get -v -d ./...
	${GO_BIN} get -u golang.org/x/lint/golint

buildmacos: dep ## Build the binary file
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 ${GO_BIN} build -a -ldflags '-extldflags "-static"' -o blchecker.macos $(PKG)

buildlinux: dep ## Build the binary file
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 ${GO_BIN} build -a -ldflags '-extldflags "-static"' -o blchecker.linux $(PKG)

compressmacos: ## Build the binary file
	${UPX} ./blchecker.macos
	${UPX} -t ./blchecker.macos

compresslinux: ## Build the binary file
	${UPX} ./blchecker.linux
	${UPX} -t ./blchecker.linux

release: clean dep lint test race buildmacos compressmacos buildlinux compresslinux

clean: ## Remove previous build
	@rm -f ./blchecker.macos
	@rm -f ./blchecker.linux

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
