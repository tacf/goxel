# Convenience commands. Uses the Go toolchain on PATH.
ARGS ?=
NAME = goxel
OUTPUT_FOLDER = bin

.DEFAULT_GOAL := help
.PHONY: help build run release test clean

help:
	@printf '%s\n' \
	  'make build      - build the game' \
	  'make run        - build and run the game' \
	  'make release    - build a stripped binary' \
	  'make test       - run unit tests' \
	  'make clean      - remove the built binary' \
	  'make run ARGS="..."  - forward command line arguments'

base:
	mkdir -p $(OUTPUT_FOLDER)

build: base
	export GOWORK=off
	go build -o $(OUTPUT_FOLDER)/$(NAME) .

run: base
	GOWORK=off go run ./cmd/goxel.go $(ARGS)

release: base
	GOWORK=off go build -ldflags="-s -w" -o $(OUTPUT_FOLDER)/$(NAME) .

test: base
	export GOWORK=off
	go test ./...

clean:
	rm -f $(OUTPUT_FOLDER)
