# This is adapted from https://habr.com/ru/post/542410/
BINARY = sync-monitor

COMMIT=$(shell git rev-parse --short HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
TAG=$(shell git describe --tags |cut -d- -f1)

LDFLAGS = -ldflags "-X main.gitTag=${TAG} -X main.gitCommit=${COMMIT} -X main.gitBranch=${BRANCH}"

.PHONY: help clean dep build install uninstall

.DEFAULT_GOAL := help

help: ## Display this help screen.
      @echo "Makefile available targets:"
      @grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  * \033[36m%-15s\033[0m %s\n", $$1, $$2}'

dep: ## Download the dependencies.
	go mod download && go get sync-monitor/cmd/sync-monitor

build: dep ## Build executable.
	mkdir -p ./bin
	CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o bin/${BINARY} ./cmd/sync-monitor

binary: build ## Alias to `build`

clean: ## Clean build directory.
	rm -f ./bin/${BINARY}
	rmdir ./bin

