BIN = gojlox

build: gen
	go build -o $(BIN)

gen:
	go generate ./...

run:
	@go run .

.PHONY: build gen run
