COVERAGE_FILE ?= coverage.out
DB_DSN ?= "postgres://admin:password@localhost:5433/merch_store_db?sslmode=disable"

.PHONY: build-app-up
build-app-up:
	@docker-compose --profile app up -d --build

.PHONY: app-up
app-up:
	@docker-compose --profile app up -d

.PHONY: infra-up
infra-up:
	@docker-compose --profile infra up -d

## test: run all tests
.PHONY: test
test:
	@go test -coverpkg='url-shortening-service/internal/application/...,url-shortening-service/internal/infrastructure/...,url-shortening-service/internal/domain' --race -count=1 -coverprofile='$(COVERAGE_FILE)' ./...
	@go tool cover -func='$(COVERAGE_FILE)' | grep ^total | tr -s '\t'

.PHONY: migrate-up
migrate-up:
	@goose -dir migrations postgres ${DB_DSN} up

.PHONY: migrate-down
migrate-down:
	@goose -dir migrations postgres ${DB_DSN} down

.PHONY: migrate-status
migrate-status:
	@goose -dir migrations postgres ${DB_DSN} status