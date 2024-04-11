DEPLOY_DIR := build

start-dev:
	@docker-compose up -d

stop-dev:
	@docker-compose down -v --remove-orphans

start-consumer:
	@OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://localhost:4317 go run cmd/consumer/main.go

start-producer:
	@OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://localhost:4317 go run cmd/producer/main.go

build:
	go build -o ${DEPLOY_DIR}/ ./cmd/*
.PHONY: build

init:
	@go mod tidy
	@go mod vendor