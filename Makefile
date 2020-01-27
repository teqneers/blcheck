PROJECT_NAME := "blcheck"
PKG := "github.com/swallo/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/...)
GO_FILES := $(shell find . -name '*.go' | grep -v _test.go)

.PHONY: all dep build clean test lint

all: build

lint: ## Lint the files
	@golint -set_exit_status ${PKG_LIST}

test: ## Run unittests
	@go test -short ${PKG_LIST}

dep: ## Get the dependencies
	@go get -v -d ./...
	@go get -u golang.org/x/lint/golint

build_macos: dep ## Build the binary file
	export GOOS=darwin 
	export GOARCH=amd64
	@go build -a -ldflags '-extldflags "-static"' -o blchecker_macos $(PKG) && upx ./blchecker_macos

build_linux: dep ## Build the binary file
	export GOOS=linux 
	export GOARCH=amd64
	@go build -a -ldflags '-extldflags "-static"' -o blchecker_linux $(PKG) && upx ./blchecker_linux

clean: ## Remove previous build
	@rm -f $(PROJECT_NAME)

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
