DEPLOY_DIR := build

start-dev:
	@docker-compose up -d

stop-dev:
	@docker-compose down -v --remove-orphans

start-consumer:
	@go run cmd/consumer/main.go

build:
	go build -o ${DEPLOY_DIR}/ ./cmd/*
.PHONY: build

init:
	@go mod tidy
	@go mod vendor