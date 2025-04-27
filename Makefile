BIN = gojlox

build: clean gen
	go build -o $(BIN)

gen:
	go generate ./...

run:
	@go run .

clean:
	$(RM) $(BIN)

.PHONY: build gen run clean
