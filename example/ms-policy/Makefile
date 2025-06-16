SHELL = /bin/sh

test:
	go test -v -cover ./internal/...
.PHONY: test

lint:
	golangci-lint run
.PHONY: lint

mockery-generate:
	go install github.com/vektra/mockery/v2@v2.52.3
	mockery --config ./mockery.yaml
.PHONY: mockery-generate

grpc-generate:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	mkdir -p ./internal/endpoints/api/backend/generated
	protoc --go-grpc_out=./internal/endpoints/api --go_out=./internal/endpoints/api internal/endpoints/api/backend/contract.proto
	mkdir -p ./internal/infrastructure/grpc/generated
	protoc --go-grpc_out=./internal/infrastructure/grpc --go_out=./internal/infrastructure/grpc internal/infrastructure/grpc/contract.proto