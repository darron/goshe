GIT_COMMIT=$(shell git rev-parse HEAD)
GOSHE_VERSION=$(shell ./version)
COMPILE_DATE=$(shell date -u +%Y%m%d.%H%M%S)
BUILD_FLAGS=-X main.CompileDate=$(COMPILE_DATE) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(GOSHE_VERSION)

all: build

deps:
	go get -u github.com/spf13/cobra
	go get -u github.com/spf13/viper
	go get -u github.com/progrium/basht
	go get -u github.com/cloudfoundry/gosigar

format:
	gofmt -w .

clean:
	rm -f bin/goshe || true

build: clean
	go build -ldflags "$(BUILD_FLAGS)" -o bin/goshe main.go

gziposx:
	gzip bin/goshe
	mv bin/goshe.gz bin/goshe-$(GOSHE_VERSION)-darwin.gz

linux: clean
	GOOS=linux GOARCH=amd64 go build -ldflags "$(BUILD_FLAGS)" -o bin/goshe main.go

gziplinux:
	gzip bin/goshe
	mv bin/goshe.gz bin/goshe-$(GOSHE_VERSION)-linux-amd64.gz

release: clean build gziposx clean linux gziplinux clean

unit:
	cd cmd && go test -v -cover

test: unit wercker

test_cleanup:
	echo "All cleaned up!"

wercker:
	basht test/tests.bash
