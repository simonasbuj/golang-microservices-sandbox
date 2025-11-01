
# run
run-frontend:
	cd front-end && go run ./cmd/web/main.go 

run-broker-service:
	cd broker-service && go run ./cmd/api

run-auth-service:
	cd auth-service && go run ./cmd/api

run-logger-service:
	cd logger-service && go run ./cmd/api

run-mail-service:
	cd mail-service && go run ./cmd/api

run-all:
	docker compose up -d --build

stop-all:
	docker compose down

# pre-commit
lint:
	cd broker-service && golangci-lint run --verbose --max-issues-per-linter=0 --max-same-issues=0
	cd auth-service && golangci-lint run --verbose --max-issues-per-linter=0 --max-same-issues=0

lint-fix:
	cd broker-service && golangci-lint run --verbose --fix
	cd auth-service && golangci-lint run --verbose --fix

.PHONY: test
test:
	cd broker-service && go test -v ./...
