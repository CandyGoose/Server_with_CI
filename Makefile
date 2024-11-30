.PHONY: build
build:
	go build -o ./bin/app ./cmd/app/main.go

.PHONY: test
test:
	go test -v ./internal/...
	go test ./... -coverprofile=cover.out

.PHONY: lint
lint:
	golangci-lint -c .golangci.yml run ./...
