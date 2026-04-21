# https://stackoverflow.com/questions/2145590/what-is-the-purpose-of-phony-in-a-makefile
.PHONY: install-tools lint

# install tools
install-tools:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/go-jet/jet/v2/cmd/jet@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

lint:
	golangci-lint run -c .golangci.yml

pull_db_image:
	docker pull postgres:18-alpine

run_db:
	docker run --name postgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:18-alpine

jet_gen:
	jet -dsn=postgresql://root:secret@localhost:5432/unravel-db?sslmode=disable -path=./db/.gen

create_migration:
	migrate create -ext sql -dir ./db/migration $1

migration_up:
	migrate -database postgres://root:secret@localhost:5432/unravel-db?sslmode=disable -path ./db/migration up

migration_down:
	migrate -database postgres://root:secret@localhost:5432/unravel-db?sslmode=disable -path ./db/migration down

create_image:
	docker build -t unravel-be .

docker_compose_up:
	docker compose up -d

start_service:
	go run cmd/main.go