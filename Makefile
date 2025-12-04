BINARY_NAME := rvtools
DB_NAME ?= duckdb.db
EXCEL_FILE ?= dallas.xlsx

.PHONY: build run

build:
	go build -o $(BINARY_NAME) .

run: build clean
	./$(BINARY_NAME) -db-path $(DB_NAME) -excel-file $(EXCEL_FILE)

clean:
	rm $(DB_NAME)
