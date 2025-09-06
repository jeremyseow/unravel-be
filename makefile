# https://stackoverflow.com/questions/2145590/what-is-the-purpose-of-phony-in-a-makefile
.PHONY: lint

lint:
	golangci-lint run -c .golangci.yml

pull_db_image:
	docker pull postgres:16-alpine

run_db:
	docker run --name postgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:16-alpine

jet_gen:
	jet -dsn=postgresql://root:secret@localhost:5432/unravel-db?sslmode=disable -path=./db/.gen

create_migration:
	migrate create -ext sql -dir ./db/migration $1
	# split up and down migrations into separate directories so that postgres image can init based on the up directory
	mv ./db/migration/*.up.sql ./db/migration/up/
	mv ./db/migration/*.down.sql ./db/migration/down/

migration_up:
	migrate -database postgres://root:secret@localhost:5432/unravel-db?sslmode=disable -path ./db/migration/up up

migration_down:
	migrate -database postgres://root:secret@localhost:5432/unravel-db?sslmode=disable -path ./db/migration/down down

create_image:
	docker build -t unravel-be .

docker_compose_up:
	docker compose up -d

start_service:
	go run cmd/main.go