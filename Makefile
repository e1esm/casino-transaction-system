PROTO_SRC=./proto/tx-manager
PROJECT_NAME=tx-manager
SERVER_OUT=services/tx-manager/src/internal
CLIENT_OUT=services/api-gateway/src/internal

protoc:
	protoc $(PROTO_SRC)/$(PROJECT_NAME).proto \
		--go_out=$(SERVER_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(SERVER_OUT) --go-grpc_opt=paths=source_relative

	protoc $(PROTO_SRC)/$(PROJECT_NAME).proto \
		--go_out=$(CLIENT_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(CLIENT_OUT) --go-grpc_opt=paths=source_relative

up:
	cd ./deployment && docker compose -p casino-transaction-system -f docker-compose-infra.yml -f docker-compose-services.yml up -d

docs:
	cd ./services/api-gateway && swag init -g ./src/cmd/main.go -o ./docs