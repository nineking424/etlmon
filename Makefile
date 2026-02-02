.PHONY: build test test-race cover clean

build:
	go build -o bin/etlmon-node ./cmd/node
	go build -o bin/etlmon-ui ./cmd/ui

test:
	go test ./...

test-race:
	go test -race ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

clean:
	rm -rf bin/ coverage.out
