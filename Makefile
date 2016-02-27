GOSHE_VERSION="0.4-dev"
GIT_COMMIT=$(shell git rev-parse HEAD)
COMPILE_DATE=$(shell date -u +%Y%m%d.%H%M%S)
BUILD_FLAGS=-X main.CompileDate=$(COMPILE_DATE) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(GOSHE_VERSION)
UNAME=$(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(shell arch)

all: build

deps:
	go get -u github.com/progrium/basht
	go get -u github.com/davecgh/go-spew/spew
	go get -u github.com/hpcloud/tail/...
	go get -u github.com/darron/goshe

format:
	gofmt -w .

clean:
	rm -f bin/goshe || true

build: clean
	go build -ldflags "$(BUILD_FLAGS)" -o bin/goshe main.go

gzip:
	gzip bin/goshe
	mv bin/goshe.gz bin/goshe-$(GOSHE_VERSION)-$(UNAME)-$(ARCH).gz

release: clean build gzip

unit:
	cd cmd && go test -v -cover

test: unit wercker

test_cleanup:
	echo "All cleaned up!"

wercker:
	basht test/tests.bash
