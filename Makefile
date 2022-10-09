-include .env.local
export

# Disable built-in rules and variables and suffixes
MAKEFLAGS += --no-builtin-rules
MAKEFLAGS += --no-builtin-variables
.SUFFIXES:

# chains common rule names to included ones
.DEFAULT_GOAL := all
.PHONY: all build clean test coverage lint run
all: build lint test coverage
build: golang-build
clean: golang-clean
test: golang-test
coverage: golang-coverage golang-coverage-html
lint: golang-lint

GO := grc go

# includes
include .make/*.mk

run:
	op run --env-file=.env.local -- go run ./cmd/siever/...
