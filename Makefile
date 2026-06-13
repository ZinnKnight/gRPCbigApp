DB_URL ?= postgres://postgres:postgres@localhost:5432/gRPCbigApp?sslmode=disable
MIGRATIONS_DIR ?= migrations
#знаю что не очень так делать, но пока вот так положил


up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f order-service market-service

proto:
	protoc 	\
    	--go_out=./protoPB --go_opt=paths=source_relative \
    	--go-grpc_out=./protoPB --go-grpc_opt=paths=source_relative \
    	--validate_out="lang=go,paths=source_relative:." \
   		-I Proto \
    	Proto/*.proto

order:
	go run ./App/OrderService/CMD

market:
	go run ./App/SpotInstrumentService/CMD

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/order-service  ./App/OrderService/OrderServiceCMD
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/market-service ./App/SpotInstrumentService/MarketServiceCMD

tidy:
	go mod tidy


migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)
