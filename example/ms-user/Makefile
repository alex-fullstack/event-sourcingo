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

start-locally:
	minikube start --driver=docker --force
	echo ${CI_POSTGRES_URL}
	kubectl delete secret user-secret -n=default --ignore-not-found \
 	&& kubectl create secret generic user-secret -n=default \
 	--from-literal=postgres-url=${CI_POSTGRES_URL} \
 	--from-literal=mongo-url=${CI_MONGO_URL} \
 	--from-literal=access-token-secret=${CI_ACCESS_TOKEN_SECRET} \
 	--from-literal=refresh-token-secret=${CI_REFRESH_TOKEN_SECRET} \
    && kubectl apply -f ms-user/charts/configmap.yaml \
    && kubectl apply -f ms-user/charts/deployment.yaml \
    && kubectl apply -f ms-user/charts/service.yaml