
up:
	docker-compose up 

proto:
	protoc --go_out=. --go_out=paths=source_relative \
		    --go-grpc_out=. --go-grpc_out=source_relative \
		    Protofiles/*.proto

order:
	go run cmd/orderService/main.go

market:
	go run cmd/marketsService/main.go

test:
	go test ./test/... -v