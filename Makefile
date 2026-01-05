.PHONY: build run test clean migrate-up migrate-down docker-up docker-down tidy

build:
	go build -o bin/server cmd/server/main.go

run:
	go run cmd/server/main.go

test:
	go test ./...

clean:
	rm -rf bin/
	go clean

tidy:
	go mod tidy
	go mod download

migrate-up:
	# TODO: добавить миграции

migrate-down:
	# TODO: добавить миграции

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-build:
	docker-compose build

docker-logs:
	docker-compose logs -f
