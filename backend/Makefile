# Tryaksh Clinic — Backend Makefile

.PHONY: run build test tidy sqlc migrate-up migrate-down migrate-create

# Run the API server
run:
	go run cmd/api/main.go

# Build the binary
build:
	go build -o bin/api cmd/api/main.go

# Run all tests
test:
	go test -v ./...

# Tidy dependencies
tidy:
	go mod tidy

# Generate sqlc code
sqlc:
	sqlc generate

# Database migrations (requires DATABASE_URL env var or .env)
migrate-up:
	migrate -path db/migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path db/migrations -database "$(DATABASE_URL)" down

migrate-create:
	migrate create -ext sql -dir db/migrations -seq $(name)
