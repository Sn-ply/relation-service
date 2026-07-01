.PHONY: run test migrate-up migrate-down docker-build setup

DB_URL ?= postgres://snaply:snaply_secret@localhost:5432/relations?sslmode=disable

setup:
	go mod tidy

run:
	DATABASE_URL=$(DB_URL) go run ./cmd/main.go

test:
	go test ./... -race -count=1

migrate-up:
	migrate -path ./migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DB_URL)" down

docker-build:
	docker build -t snaply/relation-service:latest .
