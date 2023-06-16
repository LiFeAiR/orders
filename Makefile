PROJECTNAME=$(shell basename "$(PWD)")
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard cmd/cart/main.go)
GIT=$(shell which git)
GOLANGCI_LINT=$(shell which golangci-lint run)
DOCKER_COMPOSE=$(shell which docker-compose)
export GO111MODULE=on
export GOPROXY=
export GOSUMDB=off


.PHONY: .deps-docker-pg
.deps-docker-pg:
	docker run \
	-e 'POSTGRES_USER=test' \
	-e 'POSTGRES_PASSWORD=test' \
	-p 5432:5432 \
	-d postgres:alpine \
	-c max_connections=1000


.PHONY: protoc
protoc:
	protoc -I/usr/local/include -I. \
		-I${GOPATH}/src \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway \
		--grpc-gateway_out=logtostderr=true:./api \
		--swagger_out=allow_merge=true,merge_file_name=api:./api \
		--go_out=plugins=grpc:./api ./api/*.proto