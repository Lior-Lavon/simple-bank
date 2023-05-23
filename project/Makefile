LOCAL_DB_URL=postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable
AWS_DB_URL=postgresql://root:RRjZZB39FeX7XRwXCYyn@simple-bank.cagrmci8jrd8.eu-central-1.rds.amazonaws.com:5432/simple_bank

postgres:
	docker exec -it postgres12 --network bank_network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -e POSTGRES_PASSWORD=secret -e POSTGRES_DB=simple_bank

createdb:
	docker exec -it postgres12 createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres12 dropdb simple_bank

migrateup:
	migrate -path ../db/migration -database "$(LOCAL_DB_URL)" -verbose up

aws_migrateup:
	migrate -path ../db/migration -database "$(AWS_DB_URL)" -verbose up

migrateup1:
	migrate -path ../db/migration -database "$(LOCAL_DB_URL)" -verbose up 1

migratedown:
	migrate -path ../db/migration -database "$(LOCAL_DB_URL)" -verbose down

aws_migratedown:
	migrate -path ../db/migration -database "$(AWS_DB_URL)" -verbose down

migratedown1:
	migrate -path ../db/migration -database "$(LOCAL_DB_URL)" -verbose down 1

db_docs:
	cd ../ && dbdocs build doc/db.dbml

db_schema:
	cd ../ && dbml2sql --postgres -o doc/schema.sql doc/db.dbml 

sqlc:
	cd ../db && sqlc generate

test:
	go test -v -cover ../... 

server:
	cd ../ && go run main.go

mock:
	cd ../ && mockgen -package mockdb -destination db/mock/store.go github.com/liorlavon/simplebank/db/sqlc Store 

proto:
	cd ../ && \
	rm -f pb/*.go && \
	rm -f doc/swagger/*.swagger.json && \
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
	--openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=simple_bank \
    proto/*.proto

evans:
	evans --host localhost --port 9090 -r repl

swag:
	cd ../ && \
	swag init


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
	cd ../ && go build -o ${MAIN_BINARY} .
	@echo "Done!"
