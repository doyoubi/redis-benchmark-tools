build:
	go build cmd/benchmark.go

lint:
	golangci-lint run

.PHONY: lint
