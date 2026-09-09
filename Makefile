BINARY := winding

.PHONY: run build test vet lint

run:
	go run .

build:
	go build -o bin/$(BINARY) .

test:
	go test ./...

vet:
	go vet ./...

lint:
	golangci-lint run
