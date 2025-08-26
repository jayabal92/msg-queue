PROTO_DIR=proto

.PHONY: proto build run

proto:
	protoc --go_out=paths=source_relative:. \
       --go-grpc_out=paths=source_relative:. \
       ${PROTO_DIR}/*.proto

build: proto
	go build ./...

run:
	go run ./cmd -config ./configs/config.yaml