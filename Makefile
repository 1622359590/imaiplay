.PHONY: build test run swagger docker-build docker-up docker-down docker-config

SWAG_VERSION ?= v1.16.4

build:
	mkdir -p bin
	go build -o bin/imaiplay ./cmd/server

test:
	go test ./...

run:
	go run ./cmd/server

swagger:
	go run github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION) init -g cmd/server/main.go -o docs --parseDependency --parseInternal

docker-build:
	docker compose build

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

docker-config:
	docker compose config
