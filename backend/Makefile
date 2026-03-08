.PHONY: build test test-unit test-integration test-e2e test-contract test-coverage test-load lint fmt migrate sqlc docker-up docker-down dump-routes

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

test-e2e:
	go test ./tests/e2e/... -tags=e2e -race -v

test-contract:
	bash ../scripts/test_api_contract_schemathesis.sh $(SCHEMATHESIS_ARGS)

dump-routes:
	go run ./cmd/api --dump-routes

test-clean:
	go clean -testcache
	docker container prune -f

test-load:
	k6 run tests/load/basic_load_test.js

test-load-stress:
	k6 run tests/load/stress_test.js

test-load-soak:
	k6 run tests/load/soak_test.js

test-coverage:
	go test ./... -race -coverprofile=coverage.out
	go tool cover -func=coverage.out
	@awk '/^total:/ { if ($$3+0 < 80) { print "Coverage below 80%: " $$3; exit 1 } }' coverage.out

migrate:
	go run ./cmd/migrator migrate up

migrate-down:
	go run ./cmd/migrator migrate down 1

seed-admin:
	@echo "Usage: make seed-admin EMAIL=admin@example.com PASSWORD=secret123"
	go run ./cmd/seed --email=$(EMAIL) --password=$(PASSWORD)

lint:
	golangci-lint run

fmt:
	go fmt ./...
	goimports -w .

docker-up:
	docker compose -f ../infra/docker-compose/docker-compose.dev.yml up -d

docker-down:
	docker compose -f ../infra/docker-compose/docker-compose.dev.yml down
