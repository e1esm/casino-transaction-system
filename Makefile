PROJECT_NAME := tx-manager

PROTO_DIR := ./proto/$(PROJECT_NAME)
PROTO_FILES := $(PROTO_DIR)/$(PROJECT_NAME).proto

SERVER_OUT := services/tx-manager/src/internal/proto/$(PROJECT_NAME)
CLIENT_OUT := services/api-gateway/src/internal/proto/$(PROJECT_NAME)

PROTOC_CMD = \
	protoc --proto_path=$(PROTO_DIR) $(PROTO_FILES) \
		--go_out=$(1) --go_opt=paths=source_relative \
		--go-grpc_out=$(1) --go-grpc_opt=paths=source_relative

protoc:
	@mkdir -p $(SERVER_OUT) $(CLIENT_OUT)
	$(call PROTOC_CMD,$(SERVER_OUT))
	$(call PROTOC_CMD,$(CLIENT_OUT))
	@echo "Proto generated for server and client."

up:
	cd ./deployment && docker compose -p casino-transaction-system -f docker-compose-infra.yml -f docker-compose-services.yml up -d

docs:
	docker run --rm -v $(PWD):/workspace -w /workspace ghcr.io/swaggo/swag:latest \
		init -g ./services/api-gateway/src/cmd/main.go -o ./services/api-gateway/docs

mockery:
	docker run \
	    -v $(PWD)/services/api-gateway:/api-gateway \
	    -w /api-gateway \
	    vektra/mockery:3