.PHONY: build test test-unit test-integration test-coverage test-load lint fmt migrate sqlc docker-up docker-down

build:
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker
	go build -o bin/migrator ./cmd/migrator

sqlc:
	sqlc generate

test:
	go test ./... -race -count=1

test-unit:
	go test ./internal/domain/... ./internal/application/... -race

test-integration:
	go test ./tests/integration/... -tags=integration -race

test-coverage:
	go test ./... -race -coverprofile=coverage.out
	go tool cover -func=coverage.out
	@awk '/^total:/ { if ($$3+0 < 80) { print "Coverage below 80%: " $$3; exit 1 } }' coverage.out

migrate:
	go run ./cmd/migrator migrate up

migrate-down:
	go run ./cmd/migrator migrate down 1

lint:
	golangci-lint run

fmt:
	go fmt ./...
	goimports -w .

docker-up:
	docker compose -f ../infra/docker-compose/docker-compose.dev.yml up -d

docker-down:
	docker compose -f ../infra/docker-compose/docker-compose.dev.yml down
