COVERAGE_FILE ?= coverage.out

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
	@go test -coverpkg='github.com/Lexv0lk/merch-store/internal/...,github.com/Lexv0lk/merch-store/tests/...' --race -count=1 -coverprofile='$(COVERAGE_FILE)' ./...
	@go tool cover -func='$(COVERAGE_FILE)' | grep ^total | tr -s '\t'