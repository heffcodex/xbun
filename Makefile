.PHONY: all
all: lint test

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test
test:
	go test -coverprofile cover.out ./... && go tool cover -func=cover.out

.PHONY: cover
cover: test
	go tool cover -html=cover.out && unlink cover.out
