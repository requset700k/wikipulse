BINARY := bin/server
CMD    := ./cmd/server

.PHONY: run build test lint clean

run:
	@which air > /dev/null 2>&1 && air || go run $(CMD)

build:
	go build -o $(BINARY) $(CMD)

test:
	go test ./... -v -race

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/

tidy:
	go mod tidy
