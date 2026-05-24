.PHONY: dev build test lint docker-up docker-down migrate

dev:
	air

build:
	go build -o bin/go-diamond ./cmd/server

test:
	go test ./... -v -race -cover

lint:
	golangci-lint run ./...

docker-up:
	docker-compose -f deploy/docker-compose.yml up -d

docker-down:
	docker-compose -f deploy/docker-compose.yml down

migrate:
	@for f in db/migrations/*.sql; do \
		echo "Running $$f"; \
		mysql -h127.0.0.1 -uroot -ppassword go_diamond < $$f; \
	done