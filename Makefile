# vi:syntax=make

.ONESHELL:
.DEFAULT_GOAL := help
SHELL := /bin/bash
.SHELLFLAGS = -ec

TMP_DIR?=./tmp
FUZZTIME ?= 30s
BASE_DIR=$(shell pwd)
MAKEFILE_ABSPATH := $(CURDIR)/$(word $(words $(MAKEFILE_LIST)),$(MAKEFILE_LIST))
MAKEFILE_RELPATH := $(call MAKEFILE_ABSPATH)

export GO111MODULE=on
export GOPROXY=https://proxy.golang.org
export PATH := $(BASE_DIR)/bin:$(PATH)

.PHONY: help
help: ## print help message
	@echo "Usage: make <command>"
	@echo
	@echo "Available commands are:"
	@grep -E '^\S[^:]*:.*?## .*$$' $(MAKEFILE_RELPATH) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-4s\033[36m%-30s\033[0m %s\n", "", $$1, $$2}'
	@echo

.PHONY: clean
clean:
	rm -rf $(TMP_DIR)

.PHONY: lint
lint: ## run linting
	golangci-lint run

.PHONY: test
test: ## run tests with coverage checks
	mkdir -p $(TMP_DIR)
	go test -run COMPILE_ONLY
	go test -race -cover -count=1 -coverprofile=$(TMP_DIR)/coverage.txt ./...
	@echo "Coverage report generated at $(TMP_DIR)/coverage.txt"
	@go tool cover -func=$(TMP_DIR)/coverage.txt | grep total | awk '{print "Total coverage: " $$3}'

.PHONY: fuzz
fuzz: ## run a single fuzz test (set FUZZ to match a function, e.g. make fuzz FUZZ=FuzzParse; set FUZZTIME for duration)
	go test -tags fuzz -fuzz=$(or $(FUZZ),.) -fuzztime=$(FUZZTIME)

.PHONY: fuzz-all
fuzz-all: ## run all fuzz tests sequentially (set FUZZTIME for per-test duration, e.g. make fuzz-all FUZZTIME=1m)
	@for f in $$(printf '%s\n' FuzzLexer FuzzParse FuzzRoundTrip FuzzGraphQuery FuzzQueryParse FuzzQueryExecute | shuf); do \
		case $$f in FuzzQueryParse|FuzzQueryExecute) pkg=./query ;; *) pkg=. ;; esac; \
		echo "=== fuzzing $$f for $(FUZZTIME) ==="; \
		go test -tags fuzz -fuzz=^$$f$$ -fuzztime=$(FUZZTIME) $$pkg; \
	done

.PHONY: test-integration
test-integration: build ## run integration tests (skip if tools missing)
	go test -tags integration -v ./cmd/gsl-diagram/... ./cmd/gsl-query/...

.PHONY: test-integration-strict
test-integration-strict: build ## run integration tests (fail if tools missing)
	INTEGRATION_STRICT=1 go test -tags integration -v ./cmd/gsl-diagram/... ./cmd/gsl-query/...

.PHONY: test-acceptance
test-acceptance: build ## run acceptance tests (BDD/godog feature tests)
	go test -v -tags acceptance ./test/...

.PHONY: build
build: ## build CLI tools (gsl-diagram, gsl-query)
	mkdir -p $(TMP_DIR)
	go build -o $(TMP_DIR)/gsl-diagram ./cmd/gsl-diagram
	go build -o $(TMP_DIR)/gsl-query ./cmd/gsl-query


