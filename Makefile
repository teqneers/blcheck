GO_BIN := which go
GO_LINT := which golint
UPX := which upx
PROJECT_NAME := "blcheck"
PKG := "github.com/swallo/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/...)
GO_FILES := $(shell find . -name '*.go' | grep -v _test.go)

.PHONY: all dep build clean test lint

all: build

lint: ## Lint the files
	${GO_LINT} -set_exit_status ${PKG_LIST}

test: ## Run unittests
	${GO_BIN} test -short ${PKG_LIST}

dep: ## Get the dependencies
	${GO_BIN} get -v -d ./...
	${GO_BIN} get -u golang.org/x/lint/golint

buildmacos: dep ## Build the binary file
	GOOS=darwin GOARCH=amd64 ${GO_BIN} build -a -ldflags '-extldflags "-static"' -o blchecker.macos $(PKG)

buildlinux: dep ## Build the binary file
	GOOS=linux GOARCH=amd64 ${GO_BIN} build -a -ldflags '-extldflags "-static"' -o blchecker.linux $(PKG)

compressmacos: dep ## Build the binary file
	${UPX} ./blchecker.macos

compresslinux: dep ## Build the binary file
	${UPX} ./blchecker.linux

clean: ## Remove previous build
	@rm -f $(PROJECT_NAME)

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
