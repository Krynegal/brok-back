.PHONY: test lint deps check migrate-up migrate-down run debug update-rates

# Run tests
test:
	go test -v ./...

# Install dependencies
deps:
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
lint: deps
	$(shell go env GOPATH)/bin/golangci-lint run

# Run both test and lint
check: test lint

# Run migrations up
migrate-up:
	docker run --rm -v $(PWD)/db/migrations:/migrations --network host migrate/migrate -path=/migrations -database "postgres://postgres:postgres@localhost:5433/tracker?sslmode=disable" up

# Run migrations down
migrate-down:
	docker run --rm -v $(PWD)/db/migrations:/migrations --network host migrate/migrate -path=/migrations -database "postgres://postgres:postgres@localhost:5433/tracker?sslmode=disable" down

# Update exchange rates
update-rates:
	@echo "🔄 Проверка и обновление курсов валют..."
	@curl -X POST http://localhost:8080/api/exchange-rates/update-if-needed \
		-H "Authorization: Bearer $(shell curl -s -X POST http://localhost:8080/auth/login \
			-H "Content-Type: application/json" \
			-d '{"email":"admin@example.com","password":"password"}' | jq -r '.token')" \
		|| echo "⚠️  Не удалось обновить курсы валют (возможно, сервер не запущен)"

# Run application locally
run:
	go run cmd/main.go

# Define a variable for the Delve binary
DLV := $(shell go env GOPATH)/bin/dlv

# Run application in debug mode with Delve
debug:
	@if ! [ -x "$(DLV)" ]; then \
		echo "Installing Delve..."; \
		go install github.com/go-delve/delve/cmd/dlv@latest; \
	fi
	$(DLV) debug ./cmd/main.go 