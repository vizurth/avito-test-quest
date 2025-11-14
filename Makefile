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

get-deps:
	go get -u github.com/golang-migrate/migrate

install-golangci-lint:
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0

lint:
	$(LOCAL_BIN)/golangci-lint run ./... --config .golangci.pipeline.yaml



