.PHONY: build run dev migrate-up migrate-down

build:
	go build -o bin/server ./cmd/server

run: build
	./bin/server

dev:
	go run ./cmd/server

migrate-up:
	@echo "Run migrations in order:"
	@echo "psql -d agenda -f migrations/001_create_users.sql"
	@echo "psql -d agenda -f migrations/002_create_events.sql"
	@echo "psql -d agenda -f migrations/003_create_attendance.sql"

test:
	go test -v ./...

lint:
	golangci-lint run

tidy:
	go mod tidy
