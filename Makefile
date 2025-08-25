PROTO_DIR=proto

.PHONY: gen

proto:
	protoc -I $(PROTO_DIR) --go_out=. --go-grpc_out=. $(PROTO_DIR)/*.proto

build: proto
	go build ./...

run:
	go run ./cmd/broker -config ./configs/config.yaml