postgres:
	docker exec -it postgres12 --network bank_network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -e POSTGRES_PASSWORD=secret -e POSTGRES_DB=simple_bank


createdb:
	docker exec -it postgres12 createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres12 dropdb simple_bank

migrateup:
	migrate -path ../db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up

migrateup1:
	migrate -path ../db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up 1

migratedown:
	migrate -path ../db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down

migratedown1:
	migrate -path ../db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down 1

sqlc:
	cd ../db && sqlc generate

test:
	go test -v -cover ../... 

server:
	cd ../ && go run main.go

mock:
	cd ../ && mockgen -package mockdb -destination db/mock/store.go github.com/liorlavon/simplebank/db/sqlc Store 


MAIN_BINARY=mainApp

## up: starts all containers in the background without forcing build
up:
	@echo "Starting Docker images..."
	docker-compose up -d
	@echo "Docker images started!"

## down: stop docker compose
down:
	@echo "Stopping docker compose..."
	docker-compose down
	@echo "Done!"

## up_build: stops docker-compose (if running), builds all projects and starts docker compose
up_build: build_main
	@echo "Stopping docker images (if running...)"
	docker-compose down
	@echo "Building (when required) and starting docker images..."
	docker-compose up --build -d
	@echo "Docker images built and started!"

## build_main: builds the main binary as a linux executable
build_main:
	@echo "Building main binary..."
	cd ../ && env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ${MAIN_BINARY} .
	@echo "Done!"
