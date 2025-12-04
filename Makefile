PROTO_DIR := proto
PROTO_SRC := $(wildcard $(PROTO_DIR)/*.proto)
GO_OUT := .
MODULE_NAME := github.com/skni-kod/iot-monitor-backend
BIN_DIR := bin
SERVICE ?= my-service

.PHONY: generate-proto
generate-proto:
	protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(GO_OUT) --go_opt=module=$(MODULE_NAME) \
		--go-grpc_out=$(GO_OUT) --go-grpc_opt=module=$(MODULE_NAME) \
		$(PROTO_SRC)

REBUILD_SERVICES = auth-service sensor-service api-gateway data-generation-service data-processing-service

up:
	docker-compose up --build $(REBUILD_SERVICES)

build:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/$(SERVICE) ./services/$(SERVICE)
	chmod +x $(BIN_DIR)/$(SERVICE)
