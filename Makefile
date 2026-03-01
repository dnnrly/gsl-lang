# vi:syntax=make

.ONESHELL:
.DEFAULT_GOAL := help
SHELL := /bin/bash
.SHELLFLAGS = -ec

TMP_DIR?=./tmp
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
	go test -run COMPILE_ONLY > /dev/null
	go test -race -cover -count=1 -coverprofile=$(TMP_DIR)/coverage.txt ./...
	@echo "Coverage report generated at $(TMP_DIR)/coverage.txt"
	@go tool cover -func=$(TMP_DIR)/coverage.txt | grep total | awk '{print "Total coverage: " $$3}'

.PHONY: markdown-gsl
markdown-gsl: ## validate GSL in markdown code blocks
	go test -run COMPILE_ONLY > /dev/null
	go test -count=1 -run TestMarkdownCodeBlocks

.PHONY: fuzz
fuzz: ## run fuzz tests
	go test -tags fuzz -fuzztime=2m


