.PHONY: all run local_db generate

all: generate local_db run 

run:
	go run cmd/main.go

local_db:
	cat migrations/u1_init.sql | sqlite3 /tmp/db.sqlite && cat migrations/v1_init.sql | sqlite3 /tmp/db.sqlite


generate:
	templ generate

