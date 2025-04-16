PROTO_DIR := proto
PROTO_SRC := $(wildcard $(PROTO_DIR)/*.proto)
GO_OUT := .
MODULE_NAME := github.com/skni-kod/iot-monitor-backend

.PHONY: generate-proto
generate-proto:
	protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(GO_OUT) --go_opt=module=$(MODULE_NAME) \
		--go-grpc_out=$(GO_OUT) --go-grpc_opt=module=$(MODULE_NAME) \
		$(PROTO_SRC)