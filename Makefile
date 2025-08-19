build:
	go build -o bin/pricefetcher

run: build
	./bin/pricefetcher

proto:
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
	--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	proto/*.proto

.PHONY: proto