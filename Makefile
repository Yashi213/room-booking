.PHONY: up down seed test-unit test-integration lint build

up:
	docker compose up -d --build

down:
	docker compose down

seed:
	docker compose --profile seed run --rm seed

test-unit:
	go test -count=1 ./tests/unit/...

test-integration:
	docker compose --profile test run --rm integration-tests

lint:
	golangci-lint run ./...

build:
	go build ./...
