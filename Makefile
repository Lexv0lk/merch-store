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

.PHONY: k8s-deploy
k8s-deploy:
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/secrets.yaml
	kubectl apply -f k8s/configmap.yaml
	kubectl apply -f k8s/postgres-auth/
	kubectl apply -f k8s/postgres-store/
	kubectl wait -n merch-store --for=condition=ready pod -l app=postgres-auth  --timeout=60s
	kubectl wait -n merch-store --for=condition=ready pod -l app=postgres-store --timeout=60s
	kubectl apply -f k8s/migrations/
	kubectl wait -n merch-store --for=condition=complete job/auth-migrator  --timeout=120s
	kubectl wait -n merch-store --for=condition=complete job/store-migrator --timeout=120s
	kubectl apply -f k8s/auth/
	kubectl apply -f k8s/store/
	kubectl apply -f k8s/gateway/

.PHONY: k8s-destroy
k8s-destroy:
	kubectl delete namespace merch-store

## test: run all tests
.PHONY: test
test:
	@go test -coverpkg='github.com/Lexv0lk/merch-store/internal/...,github.com/Lexv0lk/merch-store/tests/...' --race -count=1 -coverprofile='$(COVERAGE_FILE)' ./...
	@go tool cover -func='$(COVERAGE_FILE)' | grep ^total | tr -s '\t'