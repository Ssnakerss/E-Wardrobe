.PHONY: build run migrate-up migrate-down test test-race sqlc-gen lint docker-up docker-down web-copy

MODULE  := github.com/Ssnakerss/E-Wardrobe
SERVER  := ./cmd/server
CLI     := ./cmd/wardrobe
MIGRATIONS := ../../migrations

web-copy:
	cd ../web && npm run build
	rm -rf internal/webui/dist/*
	cp -r ../web/dist/* internal/webui/dist/

build: web-copy
	go build -o bin/server $(SERVER)
	go build -o bin/wardrobe $(CLI)

run:
	go run $(SERVER)

migrate-up:
	go run github.com/pressly/goose/v3/cmd/goose -dir ../../migrations up

migrate-down:
	go run github.com/pressly/goose/v3/cmd/goose -dir ../../migrations down

migrate-create:
	go run github.com/pressly/goose/v3/cmd/goose -dir ../../migrations create $(NAME) sql

test:
	go test -race -count=1 ./...

test-race:
	go test -race -count=1 -run TestRace ./...

sqlc-gen:
	sqlc generate

lint:
	golangci-lint run ./...

docker-up:
	docker compose -f ../deploy/docker-compose.dev.yaml up -d

docker-down:
	docker compose -f ../deploy/docker-compose.dev.yaml down
