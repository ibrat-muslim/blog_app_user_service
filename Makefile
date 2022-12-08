POSTGRES_HOST=host
POSTGRES_PORT=port
POSTGRES_DATABASE=database
POSTGRES_USER=user
POSTGRES_PASSWORD=password

CURRENT_DIR=$(shell pwd)

-include .env

DB_URL=postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DATABASE)?sslmode=disable

print:
	echo "$(DB_URL)"

start:
	go run cmd/main.go

migrateup:
		migrate -path migrations -database "$(DB_URL)" -verbose up

migrateup1:
		migrate -path migrations -database "$(DB_URL)" -verbose up 1

migratedown:
		migrate -path migrations -database "$(DB_URL)" -verbose down

migratedown1:
		migrate -path migrations -database "$(DB_URL)" -verbose down 1

proto-gen:
	rm -rf genproto
	./scripts/gen-proto.sh ${CURRENT_DIR}

pull-sub-module:
	git submodule update --init --recursive

update-sub-module:
	git submodule update --remote --merge


.PHONY:	start migrateup migratedown