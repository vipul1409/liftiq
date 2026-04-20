BIN := bin/reportd

.PHONY: build run test test-race vet clean

build:
	go build -o $(BIN) ./cmd/reportd

run: build
	./$(BIN)

test:
	go test ./internal/...

test-race:
	go test -race ./internal/...

vet:
	go vet ./...

clean:
	rm -rf bin/
