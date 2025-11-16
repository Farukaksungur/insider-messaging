.PHONY: build run test docker dev

build:
	go build ./...

run:
	go run ./cmd/app

dev:
	air

test:
	go test ./... -v

docker:
	docker-compose up --build
