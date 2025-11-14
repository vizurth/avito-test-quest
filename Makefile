COMPOSE_PROJECT=build/docker/docker-compose.yaml
COMPOSE_DIR=build/docker
build_and:
	cd $(COMPOSE_DIR) && docker-compose -f docker-compose.yaml build

up:
	cd $(COMPOSE_DIR) && docker-compose -f docker-compose.yaml up -d --build

down:
	cd $(COMPOSE_DIR) && docker-compose -f docker-compose.yaml down

restart: down up

LOCAL_BIN:=$(CURDIR)/bin

install-deps:
	GOBIN=$(LOCAL_BIN) go install github.com/golang-migrate/migrate
	GOBIN=$(LOCAL_BIN) go install github.com/golang/mock/mockgen@v1.6.0

get-deps:
	go get -u github.com/golang-migrate/migrate

install-golangci-lint:
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

lint: install-golangci-lint
	$(LOCAL_BIN)/golangci-lint run ./... --config .golangci.yml

lint-fix: install-golangci-lint
	$(LOCAL_BIN)/golangci-lint run ./... --config .golangci.yml --fix

format:
	gofmt -w -s ./internal ./cmd ./tests

.PHONY: cover
cover:
	go test -short -count=1 -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm coverage.out



