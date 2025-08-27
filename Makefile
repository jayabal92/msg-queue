PROTO_DIR=proto
PROJECT_NAME := msg-queue
REGISTRY     ?= jayabal92
VERSION      ?= latest

MSG_QUEUE_IMG  := $(REGISTRY)/$(PROJECT_NAME):$(VERSION)

.PHONY: proto build run

proto:
	protoc --go_out=paths=source_relative:. \
       --go-grpc_out=paths=source_relative:. \
       ${PROTO_DIR}/*.proto

build: proto
	go build ./...

run:
	go run ./cmd -config ./configs/config.yaml

.PHONY: docker
docker:
	@echo "Building docker image: $(MSG_QUEUE_IMG)"
	docker build -f Dockerfile -t $(MSG_QUEUE_IMG) .

.PHONY: push
push:
	docker push $(MSG_QUEUE_IMG)

.PHONY: minikube-load
minikube-load:
	minikube image load $(MSG_QUEUE_IMG)