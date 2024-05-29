.PHONY: all
all: lint test

.PHONY: lint
lint: _lint gci

.PHONY: gci
gci:
	GCIMODULE=`go list -m` envsubst < .golangci.gcitpl.yml | golangci-lint run -c /dev/stdin $(LINTARGS) $(LINTPATH)

.PHONY: _lint
_lint:
	golangci-lint run -c .golangci.yml $(LINTARGS) $(LINTPATH)

.PHONY: test
test:
	go test -coverprofile cover.out ./... && go tool cover -func=cover.out

.PHONY: cover
cover: test
	go tool cover -html=cover.out && unlink cover.out
