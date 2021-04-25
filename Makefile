.PHONY: tools buildRemove  test lint bench vet fmt code_style check help
GOPATH:=$(shell go env GOPATH)
TMPDIR ?= $(shell dirname $$(mktemp -u))
PACKAGE = gitlab-runner
COVER_FILE ?= $(TMPDIR)/$(PACKAGE)-coverage.out
NAMESPACE = github.com/ihippik/$(PACKAGE)
GIT_VERSION := $(shell git describe --abbrev=4 --dirty --always --tags)

$(GOLANGCI):
	GO111MODULE=off go get -u github.com/golangci/golangci-lint/cmd/golangci-lint@v1.39.0

tools: $(GOLANGCI)

build:
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-X 'main.GITVersion=$(GIT_VERSION)'" -a -installsuffix cgo ./cmd/$(PACKAGE)

test: ## Run unit (short) tests
	go test -short ./... -coverprofile=$(COVER_FILE)
	go tool cover -func=$(COVER_FILE) | grep ^total

lint: $(GOLANGCI) ## Check the project with lint
	golangci-lint run ./...

bench: ## Run benchmarks
	go test ./... -short -bench=. -run="Benchmark*"

vet: ## Check the project with vet
	go vet ./...

fmt: ## Run go fmt for the whole project
	test -z $$(for d in $$(go list -f {{.Dir}} ./...); do gofmt -e -l -w $$d/*.go; done)

check: fmt vet lint ## Run static checks (fmt, lint, vet, ...) all over the project

help: ## Print this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'