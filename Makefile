.PHONY: build test run

build:
	mkdir -p bin
	go build -o bin/imaiplay ./cmd/server

test:
	go test ./...

run:
	go run ./cmd/server
