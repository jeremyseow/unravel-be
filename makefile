# https://stackoverflow.com/questions/2145590/what-is-the-purpose-of-phony-in-a-makefile
.PHONY: lint

lint:
	golangci-lint run -c .golangci.yml

pull_db_image:
	docker pull postgres:16-alpine

run_db:
	docker run --name postgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:16-alpine

jet_gen:
	jet -dsn=postgresql://root:secret@localhost:5432/unravel-db?sslmode=disable -path=./.gen

start_service:
	go run cmd/main.go