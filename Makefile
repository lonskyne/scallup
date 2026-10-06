build:
	cd docker && docker compose up --build

proto:
	protoc --proto_path=. --go_out=. --go-grpc_out=. ./pkg/proto/*.proto
