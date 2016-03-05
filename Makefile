GOSHE_VERSION="0.4-dev"
GIT_COMMIT=$(shell git rev-parse HEAD)
COMPILE_DATE=$(shell date -u +%Y%m%d.%H%M%S)
BUILD_FLAGS=-X main.CompileDate=$(COMPILE_DATE) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(GOSHE_VERSION)
UNAME=$(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(shell arch)

all: build ## Build goshe.

deps: ## Install goshe dependencies.
	go get -u github.com/progrium/basht
	go get -u github.com/davecgh/go-spew/spew
	go get -u github.com/hpcloud/tail/...
	go get -u github.com/darron/goshe

format: ## Format the code with gofmt.
	gofmt -w .

clean: ## Remove the compiled goshe binary.
	rm -f bin/goshe || true

build: clean ## Remove the compiled binary and build.
	go build -ldflags "$(BUILD_FLAGS)" -o bin/goshe main.go

gzip: ## Gzip and rename the goshe binary according to the version, platform and architecture.
	gzip bin/goshe
	mv bin/goshe.gz bin/goshe-$(GOSHE_VERSION)-$(UNAME)-$(ARCH).gz

release: clean build gzip ## Make a complete release: clean, build and gzip.

unit: ## Run the unit tests.
	cd cmd && go test -v -cover

test: unit wercker ## Run all tests, unit and integration.

test_cleanup: ## Cleanup after the test run.
	echo "All cleaned up!"

wercker: ## Run the integration tests.
	basht test/tests.bash

.PHONY: help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
